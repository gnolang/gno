// Package multisig is a factory for M-of-N treasuries.
//
// Anyone may create a safe. Each safe's coins live at its own
// sub-realm address. Signers propose, approve, revoke, cancel, and
// execute. v0 is unaudited.
package multisig

import (
	"chain"
	"chain/banker"
	"chain/runtime/unsafe"
	"strconv"
	"time"

	"gno.land/p/nt/grc20/v0"
	msig "gno.land/p/nt/multisig/v0"

	"gno.land/r/nt/grc20reg/v0"
)

// maxSafes bounds the factory. Create is permissionless, and every
// safe is reachable from this list, so an open-ended list is a
// storage sink. 1024 is a v0 ceiling, not a product decision about
// how many treasuries the chain should have.
const maxSafes = 1024

var (
	safes     []*msig.Safe
	realmPath string
)

func init(cur realm) {
	realmPath = cur.PkgPath()
}

// Create opens a safe. signers is a comma-separated list of bech32
// addresses. threshold is how many of them must approve. delay is the
// timelock in seconds for proposals created later; 0 executes as soon
// as the threshold is met.
func Create(cur realm, name, signers string, threshold int, delay int64) int64 {
	if len(safes) >= maxSafes {
		panic(msig.ErrTooManySafes.Error())
	}
	caller := cur.Previous().Address()
	s, err := msig.New(cur.PkgPath(), cur.Address(), caller, name, signers, threshold, delay)
	must(err)
	id := int64(len(safes) + 1)
	_, err = s.Bind(id)
	must(err)
	safes = append(safes, s)
	return id
}

// Deposit forwards every coin in this message's -send envelope to the
// safe so nothing can be stranded at the factory address. The call must
// come directly from the signing user. Realms that already hold coins
// send them straight to SafeAddress instead.
func Deposit(cur realm, safeID int64) {
	if !cur.Previous().IsUserCall() {
		panic(msig.ErrDeposit.Error())
	}
	s := mustSafe(safeID)
	sent := unsafe.OriginSend()
	if len(sent) == 0 {
		panic(msig.ErrAmount.Error())
	}
	for _, c := range sent {
		if c.Amount <= 0 {
			panic(msig.ErrAmount.Error())
		}
		if !msig.ValidDenom(c.Denom) {
			panic(msig.ErrDenom.Error())
		}
	}
	b := banker.NewBanker(banker.BankerTypeOriginSend, cur)
	b.SendCoins(cur.Address(), s.Address(), sent)
	from := cur.Previous().Address().String()
	sid := strconv.FormatInt(safeID, 10)
	for _, c := range sent {
		chain.Emit(
			"SafeDeposit",
			"safe", sid,
			"from", from,
			"denom", c.Denom,
			"amount", strconv.FormatInt(c.Amount, 10),
		)
	}
}

// ProposeCoin queues a native-coin payment. expiresAt is a unix second.
func ProposeCoin(cur realm, safeID int64, to address, denom string, amount, expiresAt int64, memo string) int64 {
	id, err := mustSafe(safeID).ProposeCoin(0, cur, to, denom, amount, expiresAt, memo, nowUnix())
	must(err)
	return id
}

// ProposeGRC20 queues a transfer of the token registered at tokenKey.
func ProposeGRC20(cur realm, safeID int64, to address, tokenKey string, amount, expiresAt int64, memo string) int64 {
	id, err := mustSafe(safeID).ProposeGRC20(0, cur, to, tokenKey, amount, expiresAt, memo, nowUnix())
	must(err)
	return id
}

// ProposeCall queues a one-shot authorization. Execute does not call
// targetPkg. That realm pulls the payload with Consume.
func ProposeCall(cur realm, safeID int64, targetPkg, payload string, expiresAt int64, memo string) int64 {
	id, err := mustSafe(safeID).ProposeCall(0, cur, targetPkg, payload, expiresAt, memo, nowUnix())
	must(err)
	return id
}

// ProposeAddSigner queues a new signer.
func ProposeAddSigner(cur realm, safeID int64, signer address, expiresAt int64, memo string) int64 {
	id, err := mustSafe(safeID).ProposeAddSigner(0, cur, signer, expiresAt, memo, nowUnix())
	must(err)
	return id
}

// ProposeRemoveSigner queues a signer removal. Outstanding approvals
// from that signer stop counting, including if the signer is added back.
func ProposeRemoveSigner(cur realm, safeID int64, signer address, expiresAt int64, memo string) int64 {
	id, err := mustSafe(safeID).ProposeRemoveSigner(0, cur, signer, expiresAt, memo, nowUnix())
	must(err)
	return id
}

// ProposeThreshold queues a new threshold.
func ProposeThreshold(cur realm, safeID int64, threshold int, expiresAt int64, memo string) int64 {
	id, err := mustSafe(safeID).ProposeThreshold(0, cur, threshold, expiresAt, memo, nowUnix())
	must(err)
	return id
}

// ProposeDelay queues the timelock used by proposals created afterwards.
func ProposeDelay(cur realm, safeID int64, delay, expiresAt int64, memo string) int64 {
	id, err := mustSafe(safeID).ProposeDelay(0, cur, delay, expiresAt, memo, nowUnix())
	must(err)
	return id
}

// ProposePause queues a pause. Spends wait; signer and pause changes do not.
func ProposePause(cur realm, safeID int64, expiresAt int64, memo string) int64 {
	id, err := mustSafe(safeID).ProposePause(0, cur, expiresAt, memo, nowUnix())
	must(err)
	return id
}

// ProposeUnpause queues an unpause.
func ProposeUnpause(cur realm, safeID int64, expiresAt int64, memo string) int64 {
	id, err := mustSafe(safeID).ProposeUnpause(0, cur, expiresAt, memo, nowUnix())
	must(err)
	return id
}

// Approve adds the caller's vote. A second vote from the same signer fails.
func Approve(cur realm, safeID, propID int64) {
	must(mustSafe(safeID).Approve(0, cur, propID, nowUnix()))
}

// Revoke removes the caller's vote and restarts the timelock if the
// proposal falls back under the threshold.
func Revoke(cur realm, safeID, propID int64) {
	must(mustSafe(safeID).Revoke(0, cur, propID, nowUnix()))
}

// Cancel drops a proposal that has not yet reached the threshold.
func Cancel(cur realm, safeID, propID int64) {
	must(mustSafe(safeID).Cancel(0, cur, propID, nowUnix()))
}

// Execute performs a proposal that has enough approvals and whose
// timelock has elapsed. It runs at most once.
func Execute(cur realm, safeID, propID int64) {
	s := mustSafe(safeID)
	now := nowUnix()
	eff, err := s.CheckExecute(0, cur, propID, now)
	must(err)
	must(canAfford(s, eff))
	eff, err = s.Execute(0, cur, propID, now)
	must(err)
	apply(cur, s, eff)
}

func canAfford(s *msig.Safe, eff msig.Effect) error {
	switch eff.Kind {
	case msig.KindCoin:
		if !msig.ValidDenom(eff.Denom) {
			return msig.ErrDenom
		}
		if banker.NewReadonlyBanker().GetCoin(s.Address(), eff.Denom) < eff.Amount {
			return msig.ErrInsufficientBalance
		}
	case msig.KindGRC20:
		tok := grc20reg.Get(eff.TokenKey)
		if tok == nil {
			return msig.ErrToken
		}
		if tok.BalanceOf(s.Address()) < eff.Amount {
			return msig.ErrInsufficientBalance
		}
	}
	return nil
}

// Consume returns the payload of an executed call proposal, once.
// Only a direct call from targetPkg succeeds.
func Consume(cur realm, safeID, propID int64) string {
	payload, err := mustSafe(safeID).Consume(0, cur, propID)
	must(err)
	return payload
}

// WatchDenom shows denom on the safe's page. Any signer may call it.
func WatchDenom(cur realm, safeID int64, denom string) {
	must(mustSafe(safeID).WatchDenom(0, cur, denom))
}

// WatchToken shows a GRC20 registry key on the safe's page.
func WatchToken(cur realm, safeID int64, tokenKey string) {
	must(mustSafe(safeID).WatchToken(0, cur, tokenKey))
}

// UnwatchDenom removes denom from the page.
func UnwatchDenom(cur realm, safeID int64, denom string) {
	must(mustSafe(safeID).UnwatchDenom(0, cur, denom))
}

// UnwatchToken removes a token key from the page.
func UnwatchToken(cur realm, safeID int64, tokenKey string) {
	must(mustSafe(safeID).UnwatchToken(0, cur, tokenKey))
}

func apply(cur realm, s *msig.Safe, eff msig.Effect) {
	switch eff.Kind {
	case msig.KindCoin:
		applyCoin(cur, s, eff)
	case msig.KindGRC20:
		applyGRC20(cur, s, eff)
	}
}

func applyCoin(cur realm, s *msig.Safe, eff msig.Effect) {
	if eff.Amount <= 0 {
		panic(msig.ErrAmount.Error())
	}
	if !msig.ValidDenom(eff.Denom) {
		panic(msig.ErrDenom.Error())
	}
	sub := cur.Sub(s.Subpath())
	if sub.Address() != s.Address() {
		panic("safe address does not match its sub-realm")
	}
	if banker.NewReadonlyBanker().GetCoin(sub.Address(), eff.Denom) < eff.Amount {
		panic(msig.ErrInsufficientBalance.Error())
	}
	b := banker.NewBanker(banker.BankerTypeRealmSend, sub)
	b.SendCoins(sub.Address(), eff.To, chain.NewCoins(chain.NewCoin(eff.Denom, eff.Amount)))
}

func applyGRC20(cur realm, s *msig.Safe, eff msig.Effect) {
	tok := grc20reg.Get(eff.TokenKey)
	if tok == nil {
		panic(msig.ErrToken.Error())
	}
	sub := cur.Sub(s.Subpath())
	if sub.Address() != s.Address() {
		panic("safe address does not match its sub-realm")
	}
	if tok.BalanceOf(sub.Address()) < eff.Amount {
		panic(msig.ErrInsufficientBalance.Error())
	}
	teller := tok.RealmTeller(0, sub)
	if !grc20.IsCanonicalTeller(teller) {
		panic(msig.ErrToken.Error())
	}
	if err := teller.Transfer(0, sub, eff.To, eff.Amount); err != nil {
		panic(err.Error())
	}
}

func mustSafe(id int64) *msig.Safe {
	if id <= 0 || id > int64(len(safes)) || safes[int(id-1)] == nil {
		panic(msig.ErrUnknownSafe.Error())
	}
	return safes[int(id-1)]
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func nowUnix() int64 { return time.Now().Unix() }
