// PKGPATH: gno.land/r/nt/multisig/v0/filetests/z_adversary

package z_adversary

import (
	"chain"
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/grc20/v0"
	"gno.land/p/nt/testutils/v0"

	"gno.land/r/nt/grc20reg/v0"
	"gno.land/r/nt/multisig/v0"
)

const t0 int64 = 1_700_000_000

// hits counts callbacks. A teller transfer and an execute must not
// enter this realm.
var hits int

func Notify(cur realm) {
	hits++
}

func main(cur realm) {
	setTime(t0)
	tok, led := grc20.NewToken("Adversary", "ADV", 0, 1, cur)
	key := grc20reg.Register(cross(cur), tok, "")

	alice := testutils.TestAddress("alice")
	pkg := cur.PkgPath()
	me := cur.Address()
	asUser(t0, alice)
	id := multisig.Create(cross(cur), "adv", alice.String(), 1, 0)
	if err := led.Mint(multisig.SafeAddress(id), 3); err != nil {
		panic(err)
	}
	testing.IssueCoins(multisig.SafeAddress(id), chain.NewCoins(chain.NewCoin("ugnot", 4)))

	// Relaying published realm: OriginCaller is alice, but cur.Previous()
	// is this realm (pkg), so Deposit rejects IsUserCall and signer-gated
	// calls reject because this realm's address is not a signer.
	asCode(t0, alice, pkg)
	expectPanic("direct user call", func() { multisig.Deposit(cross(cur), id) })
	expectPanic("not a signer", func() { multisig.Approve(cross(cur), id, 1) })
	expectPanic("not a signer", func() { multisig.Execute(cross(cur), id, 1) })

	// A different realm, even with the user's origin, is not the signer.
	asCode(t0, alice, "gno.land/r/demo/evil")
	expectPanic("not a signer", func() { multisig.ProposeCoin(cross(cur), id, me, "ugnot", 1, t0+1000, "") })

	asUser(t0, alice)
	coin := multisig.ProposeCoin(cross(cur), id, me, "ugnot", 4, t0+1000, "")
	call := multisig.ProposeCall(cross(cur), id, pkg, "secret", t0+1000, "")
	tokProp := multisig.ProposeGRC20(cross(cur), id, me, key, 3, t0+1000, "")
	multisig.Approve(cross(cur), id, coin)
	multisig.Approve(cross(cur), id, call)
	multisig.Approve(cross(cur), id, tokProp)

	if hits != 0 {
		panic("proposed into the target")
	}
	multisig.Execute(cross(cur), id, coin)
	multisig.Execute(cross(cur), id, tokProp)
	multisig.Execute(cross(cur), id, call)
	if hits != 0 {
		panic("execute entered the recipient or the target")
	}
	if tok.BalanceOf(me) != 3 || tok.BalanceOf(multisig.SafeAddress(id)) != 0 {
		panic("token balances")
	}

	// Impersonation: origin is alice, caller realm is evil.
	asCode(t0, alice, "gno.land/r/demo/evil")
	expectPanic("unauthorized", func() { multisig.Consume(cross(cur), id, call) })

	// A published realm explicitly listed as a signer can act on its own behalf.
	asUser(t0, alice)
	realmSafe := multisig.Create(cross(cur), "dao-safe", me.String(), 1, 0)
	asCode(t0, alice, pkg)
	rprop := multisig.ProposeDelay(cross(cur), realmSafe, 5, t0+1000, "")
	multisig.Approve(cross(cur), realmSafe, rprop)
	multisig.Execute(cross(cur), realmSafe, rprop)
	if multisig.SafeDelay(realmSafe) != 5 {
		panic("realm signer delay")
	}

	ctx := testing.GetContext()
	ctx.Time = time.Unix(t0, 0)
	ctx.CurrentRealm = testing.NewCodeRealm(pkg)
	ctx.OriginCaller = alice
	testing.SetContext(ctx)
	payload := multisig.Consume(cross(cur), id, call)
	if payload != "secret" {
		panic("payload")
	}
	expectPanic("already consumed", func() { multisig.Consume(cross(cur), id, call) })
	expectPanic("not a signer", func() { multisig.Execute(cross(cur), id, coin) })
	if multisig.SafeNonce(id) != 3 {
		panic("nonce")
	}
	Notify(cross(cur))
	if hits != 1 {
		panic("notify")
	}
	println("ok")
}

func setTime(unix int64) {
	ctx := testing.GetContext()
	ctx.Time = time.Unix(unix, 0)
	testing.SetContext(ctx)
}

func asUser(unix int64, user address) {
	ctx := testing.GetContext()
	ctx.Time = time.Unix(unix, 0)
	ctx.CurrentRealm = testing.NewUserRealm(user)
	ctx.OriginCaller = user
	testing.SetContext(ctx)
}

func asCode(unix int64, origin address, pkgPath string) {
	ctx := testing.GetContext()
	ctx.Time = time.Unix(unix, 0)
	ctx.CurrentRealm = testing.NewCodeRealm(pkgPath)
	ctx.OriginCaller = origin
	testing.SetContext(ctx)
}

func expectPanic(substr string, fn func()) {
	r := revive(fn)
	if r == nil {
		panic("expected panic: " + substr)
	}
	msg := ""
	switch v := r.(type) {
	case string:
		msg = v
	case error:
		msg = v.Error()
	default:
		msg = "unknown"
	}
	if !strings.Contains(msg, substr) {
		panic("panic " + msg + " missing " + substr)
	}
}

// Output:
// ok
