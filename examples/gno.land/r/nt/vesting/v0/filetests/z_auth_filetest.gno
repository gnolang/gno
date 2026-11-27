// PKGPATH: gno.land/r/nt/vesting/v0/filetests/z_auth

package z_auth

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

// OriginCaller is not the beneficiary. A code realm that sets the
// origin caller to the beneficiary still presents its own address.
func main(cur realm) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	tok, led := grc20.NewToken("Auth", "AUTH", 0, 1, cur)
	key := grc20reg.Register(cross(cur), tok, "")
	factory := chain.PackageAddress("gno.land/r/nt/vesting/v0")
	must(led.Mint(alice, 20))
	must(led.Approve(alice, factory, 20))

	as(t0, alice)
	id := vesting.CreateGRC20(cross(cur), alice, key, 20, t0, 0, 10, 0, 1, true, "auth")

	// This package calls, while pretending the signer is alice.
	impersonate(t0+10, alice)
	expectPanic("not the beneficiary", func() { vesting.Claim(cross(cur), id) })
	if vesting.ClaimedOf(id) != 0 {
		panic("impersonation claimed")
	}

	as(t0+10, bob)
	expectPanic("not the beneficiary", func() { vesting.Claim(cross(cur), id) })
	expectPanic("not the admin", func() { vesting.Revoke(cross(cur), id) })

	as(t0+10, alice)
	vesting.Claim(cross(cur), id)
	if vesting.ClaimedOf(id) != 20 {
		panic("alice claim")
	}
	if tok.BalanceOf(alice) != 20 {
		panic("alice balance")
	}
	println("ok")
}

func impersonate(unix int64, user address) {
	ctx := testing.GetContext()
	ctx.Time = time.Unix(unix, 0)
	ctx.OriginCaller = user
	ctx.CurrentRealm = testing.NewCodeRealm("gno.land/r/nt/vesting/v0/filetests/z_auth")
	testing.SetContext(ctx)
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
