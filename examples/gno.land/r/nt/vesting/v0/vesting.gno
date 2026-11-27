// Package vesting is a factory for cliff-and-vest grants.
//
// Anyone may create a grant. Each grant's coins live at its own
// sub-realm address. The beneficiary claims, and may name a new
// beneficiary. A revocable grant's admin may revoke, which returns
// the unvested part and leaves the vested part claimable.
//
// Every mutation checks Grant.Invariant before it returns. A send is
// accepted only when the source lost exactly the amount and the
// destination gained exactly the amount. This factory does not take a
// caller-supplied teller and does not call anything except the bank
// and the canonical GRC20 token. v0 is unaudited.
package vesting

import (
	"chain"
	"chain/banker"
	"chain/runtime/unsafe"
	"math/overflow"
	"time"

	"gno.land/p/nt/grc20/v0"
	vst "gno.land/p/nt/vesting/v0"

	"gno.land/r/nt/grc20reg/v0"
)

var (
	grants    []*vst.Grant
	realmPath string
	busy      bool
)

func init(cur realm) {
	realmPath = cur.PkgPath()
}

// CreateCoin opens a grant funded by this message's -send envelope.
// The envelope must be exactly denom/total and nothing else. The call
// must come from the signing user: a maketx run script can spend the
// envelope before this function sees it.
func CreateCoin(cur realm, beneficiary address, denom string, total, start, cliff, duration, period, curve int64, revocable bool, label string) int64 {
	unlock := guard()
	defer unlock()
	if !cur.Previous().IsUserCall() {
		panic(vst.ErrUserCall.Error())
	}
	// Pair the user-call check with the envelope. IsUserCall means the
	// coins landed on this realm, so the envelope and the receipt agree.
	sent := unsafe.OriginSend()
	if len(sent) != 1 || sent[0].Denom != denom || sent[0].Amount != total {
		panic(vst.ErrPayment.Error())
	}
	now := mustNow()
	g := newGrant(cur, cur.Previous().Address(), beneficiary, vst.Asset{Kind: vst.AssetCoin, Denom: denom}, total, start, cliff, duration, period, curve, revocable, label)
	id, addr := bindNext(g)
	bkr := banker.NewReadonlyBanker()
	srcBefore := bkr.GetCoin(cur.Address(), denom)
	dstBefore := bkr.GetCoin(addr, denom)
	if _, ok := overflow.Add64(dstBefore, total); !ok {
		panic(vst.ErrMath.Error())
	}
	b := banker.NewBanker(banker.BankerTypeOriginSend, cur)
	b.SendCoins(cur.Address(), addr, chain.NewCoins(chain.NewCoin(denom, total)))
	// The factory may already hold this denom. Only `total` may leave,
	// and the grant address must gain that same `total`.
	if !moved(srcBefore, dstBefore, total, bkr.GetCoin(cur.Address(), denom), bkr.GetCoin(addr, denom)) {
		panic(vst.ErrFunding.Error())
	}
	grants = append(grants, g)
	commit(g, now, true)
	return id
}

// CreateGRC20 opens a grant funded by an allowance. The caller must
// already have approved this factory's address, not the grant address,
// for total units of the token registered at tokenKey.
func CreateGRC20(cur realm, beneficiary address, tokenKey string, total, start, cliff, duration, period, curve int64, revocable bool, label string) int64 {
	unlock := guard()
	defer unlock()
	now := mustNow()
	tok := grc20reg.Get(tokenKey)
	if tok == nil {
		panic(vst.ErrToken.Error())
	}
	caller := cur.Previous().Address()
	g := newGrant(cur, caller, beneficiary, vst.Asset{Kind: vst.AssetGRC20, Token: tokenKey}, total, start, cliff, duration, period, curve, revocable, label)
	id, addr := bindNext(g)
	// RealmTeller is a method on the concrete *grc20.Token the registry
	// returns. The teller is still checked: this factory never accepts a
	// Teller value from a caller.
	teller := tok.RealmTeller(0, cur)
	if !grc20.IsCanonicalTeller(teller) {
		panic(vst.ErrToken.Error())
	}
	srcBefore := tok.BalanceOf(caller)
	dstBefore := tok.BalanceOf(addr)
	if _, ok := overflow.Add64(dstBefore, total); !ok {
		panic(vst.ErrMath.Error())
	}
	if err := teller.TransferFrom(0, cur, caller, addr, total); err != nil {
		panic(err.Error())
	}
	if !moved(srcBefore, dstBefore, total, tok.BalanceOf(caller), tok.BalanceOf(addr)) {
		panic(vst.ErrFunding.Error())
	}
	grants = append(grants, g)
	commit(g, now, true)
	return id
}

// Claim sends the beneficiary everything that has vested and not yet
// been claimed. The caller is cur.Previous().Address(), which is the
// signing user on maketx call and the calling realm on a cross-call.
func Claim(cur realm, id int64) {
	unlock := guard()
	defer unlock()
	g := mustGrant(id)
	caller := cur.Previous().Address()
	now := nowUnix()
	if caller != g.Beneficiary() {
		panic(vst.ErrNotBeneficiary.Error())
	}
	rel, err := g.Releasable(now)
	must(err)
	if rel == 0 {
		panic(vst.ErrNothing.Error())
	}
	checkFunded(g, rel)
	got, err := g.Claim(caller, now)
	must(err)
	if got != rel {
		panic(vst.ErrMath.Error())
	}
	pay(cur, g, caller, got)
	commit(g, now, true)
}

// Revoke returns the unvested balance to the admin and leaves the
// vested balance claimable. Only a revocable grant's admin may call it,
// and only once.
func Revoke(cur realm, id int64) {
	unlock := guard()
	defer unlock()
	g := mustGrant(id)
	caller := cur.Previous().Address()
	now := nowUnix()
	if caller != g.Admin() {
		panic(vst.ErrNotAdmin.Error())
	}
	back, err := g.Reclaimable(now)
	must(err)
	checkFunded(g, back)
	got, err := g.Revoke(caller, now)
	must(err)
	if got != back {
		panic(vst.ErrMath.Error())
	}
	if got > 0 {
		pay(cur, g, caller, got)
	}
	commit(g, now, true)
}

// SetBeneficiary moves the claim to next. Only the current beneficiary
// may call it. The admin cannot redirect a grant.
func SetBeneficiary(cur realm, id int64, next address) {
	unlock := guard()
	defer unlock()
	g := mustGrant(id)
	now := mustNow()
	must(g.SetBeneficiary(cur.Previous().Address(), next))
	commit(g, now, false)
}

// TransferAdmin moves the revoke right. On an irrevocable grant the
// record moves and Revoke still fails.
func TransferAdmin(cur realm, id int64, next address) {
	unlock := guard()
	defer unlock()
	g := mustGrant(id)
	now := mustNow()
	must(g.TransferAdmin(cur.Previous().Address(), next))
	commit(g, now, false)
}

func newGrant(cur realm, admin, beneficiary address, asset vst.Asset, total, start, cliff, duration, period, curve int64, revocable bool, label string) *vst.Grant {
	c, err := vst.ParseCurve(curve)
	must(err)
	g, err := vst.New(
		cur.PkgPath(),
		cur.Address(),
		admin,
		beneficiary,
		asset,
		vst.Terms{Start: start, Cliff: cliff, Duration: duration, Period: period, Total: total, Curve: c},
		revocable,
		label,
	)
	must(err)
	return g
}

func bindNext(g *vst.Grant) (int64, address) {
	id, ok := overflow.Add64(int64(len(grants)), 1)
	if !ok || id <= 0 {
		panic(vst.ErrFull.Error())
	}
	addr, err := g.Bind(id)
	must(err)
	return id, addr
}

func funded(before, amount, after int64) bool {
	want, ok := overflow.Add64(before, amount)
	return ok && after == want
}

func spent(before, amount, after int64) bool {
	want, ok := overflow.Sub64(before, amount)
	return ok && after == want
}

// checkFunded verifies the grant address covers the full outstanding
// liability (and at least need) before any books move or coins are sent.
// Because Claim and Revoke reduce both balance and outstanding by the
// same amount, checking this upfront prevents partial-drain mutations
// from reaching commit.
func checkFunded(g *vst.Grant, need int64) {
	if need < 0 {
		panic(vst.ErrMath.Error())
	}
	out, err := g.Outstanding()
	must(err)
	if need > out {
		out = need
	}
	if live(g) < out {
		panic(vst.ErrBalance.Error())
	}
}

func pay(cur realm, g *vst.Grant, to address, amount int64) {
	if amount <= 0 {
		panic(vst.ErrMath.Error())
	}
	sub := cur.Sub(g.Subpath())
	if sub.Address() != g.Address() {
		panic("grant address does not match its sub-realm")
	}
	if to == sub.Address() || to == cur.Address() {
		panic(vst.ErrSelf.Error())
	}
	switch g.Asset().Kind {
	case vst.AssetCoin:
		bkr := banker.NewReadonlyBanker()
		srcBefore := bkr.GetCoin(sub.Address(), g.Asset().Denom)
		dstBefore := bkr.GetCoin(to, g.Asset().Denom)
		if srcBefore < amount {
			panic(vst.ErrBalance.Error())
		}
		if _, ok := overflow.Add64(dstBefore, amount); !ok {
			panic(vst.ErrMath.Error())
		}
		b := banker.NewBanker(banker.BankerTypeRealmSend, sub)
		b.SendCoins(sub.Address(), to, chain.NewCoins(chain.NewCoin(g.Asset().Denom, amount)))
		if !moved(srcBefore, dstBefore, amount, bkr.GetCoin(sub.Address(), g.Asset().Denom), bkr.GetCoin(to, g.Asset().Denom)) {
			panic(vst.ErrFunding.Error())
		}
	case vst.AssetGRC20:
		tok := grc20reg.Get(g.Asset().Token)
		if tok == nil {
			panic(vst.ErrToken.Error())
		}
		srcBefore := tok.BalanceOf(sub.Address())
		dstBefore := tok.BalanceOf(to)
		if srcBefore < amount {
			panic(vst.ErrBalance.Error())
		}
		if _, ok := overflow.Add64(dstBefore, amount); !ok {
			panic(vst.ErrMath.Error())
		}
		teller := tok.RealmTeller(0, sub)
		if !grc20.IsCanonicalTeller(teller) {
			panic(vst.ErrToken.Error())
		}
		if err := teller.Transfer(0, sub, to, amount); err != nil {
			panic(err.Error())
		}
		if !moved(srcBefore, dstBefore, amount, tok.BalanceOf(sub.Address()), tok.BalanceOf(to)) {
			panic(vst.ErrFunding.Error())
		}
	default:
		panic(vst.ErrAsset.Error())
	}
}

func moved(srcBefore, dstBefore, amount, srcAfter, dstAfter int64) bool {
	return spent(srcBefore, amount, srcAfter) && funded(dstBefore, amount, dstAfter)
}

// commit re-checks the books after a mutation. funds also requires the
// grant address to still hold at least what the books say is outstanding.
// A donation may leave more. A shortfall panics, and the transaction
// reverts the mutation.
func commit(g *vst.Grant, now int64, funds bool) {
	must(g.Invariant(now))
	if !funds {
		return
	}
	out, err := g.Outstanding()
	must(err)
	if live(g) < out {
		panic(vst.ErrFunding.Error())
	}
}

func live(g *vst.Grant) int64 {
	bal, err := assetBalance(g.Asset(), g.Address())
	must(err)
	return bal
}

func assetBalance(a vst.Asset, addr address) (int64, error) {
	switch a.Kind {
	case vst.AssetCoin:
		return banker.NewReadonlyBanker().GetCoin(addr, a.Denom), nil
	case vst.AssetGRC20:
		tok := grc20reg.Get(a.Token)
		if tok == nil {
			return 0, vst.ErrToken
		}
		return tok.BalanceOf(addr), nil
	default:
		return 0, vst.ErrAsset
	}
}

func mustGrant(id int64) *vst.Grant {
	if id <= 0 || id > int64(len(grants)) {
		panic(vst.ErrUnknown.Error())
	}
	g := grants[id-1]
	if g == nil || g.ID() != id {
		panic(vst.ErrUnknown.Error())
	}
	return g
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func guard() func() {
	if busy {
		panic(vst.ErrBusy.Error())
	}
	busy = true
	return func() { busy = false }
}

func nowUnix() int64 { return time.Now().Unix() }

func mustNow() int64 {
	now := nowUnix()
	if now < 0 {
		panic(vst.ErrTime.Error())
	}
	return now
}
