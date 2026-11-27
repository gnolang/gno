// PKGPATH: gno.land/r/nt/vesting/v0/filetests/z_grc20

package z_grc20

import (
	"chain"
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/grc20/v0"
	"gno.land/p/nt/testutils/v0"

	"gno.land/r/nt/grc20reg/v0"
	"gno.land/r/nt/vesting/v0"
)

const t0 int64 = 1_700_000_000

func main(cur realm) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	tok, led := grc20.NewToken("Grant", "GRNT", 0, 1, cur)
	key := grc20reg.Register(cross(cur), tok, "")
	factory := chain.PackageAddress("gno.land/r/nt/vesting/v0")
	must(led.Mint(alice, 100))
	must(led.Mint(bob, 40))
	must(led.Approve(alice, factory, 100))
	must(led.Approve(bob, factory, 40))

	as(t0, alice)
	expectPanic("invalid token", func() {
		vesting.CreateGRC20(cross(cur), alice, "missing.token", 1, t0, 0, 1, 0, 1, true, "")
	})

	as(t0, alice)
	idA := vesting.CreateGRC20(cross(cur), bob, key, 100, t0, 0, 10, 0, 1, true, "alice-admin")
	as(t0, bob)
	idB := vesting.CreateGRC20(cross(cur), bob, key, 40, t0, 10, 10, 0, 1, false, "bob")

	as(t0+5, bob)
	expectPanic("nothing to release", func() { vesting.Claim(cross(cur), idB) })
	vesting.Claim(cross(cur), idA)
	if vesting.ClaimedOf(idA) != 50 || vesting.BalanceOf(idB) != 40 {
		panic("isolation")
	}

	as(t0+5, alice)
	vesting.Revoke(cross(cur), idA)
	expectPanic("already revoked", func() { vesting.Revoke(cross(cur), idA) })
	if vesting.ReturnedOf(idA) != 50 {
		panic("returned")
	}
	if tok.BalanceOf(alice) != 50 {
		panic("admin balance")
	}

	as(t0+10, bob)
	expectPanic("nothing to release", func() { vesting.Claim(cross(cur), idA) })
	expectPanic("irrevocable", func() { vesting.Revoke(cross(cur), idB) })
	vesting.Claim(cross(cur), idB)
	if vesting.BalanceOf(idA) != 0 || vesting.BalanceOf(idB) != 0 {
		panic("drained books")
	}
	if tok.BalanceOf(bob) != 90 {
		panic("bob balance")
	}
	println("ok")
}

func as(unix int64, user address) {
	ctx := testing.GetContext()
	ctx.Time = time.Unix(unix, 0)
	ctx.CurrentRealm = testing.NewUserRealm(user)
	ctx.OriginCaller = user
	testing.SetContext(ctx)
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
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
