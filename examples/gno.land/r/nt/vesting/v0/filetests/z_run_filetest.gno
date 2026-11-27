// PKGPATH: gno.land/r/nt/vesting/v0/filetests/z_run

package z_run

import (
	"chain"
	"chain/runtime"
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/grc20/v0"
	"gno.land/p/nt/testutils/v0"

	"gno.land/r/nt/grc20reg/v0"
	"gno.land/r/nt/vesting/v0"
)

const t0 int64 = 1_700_000_000

// maketx run is the user for a claim and is not a direct call for a
// coin envelope. Another user's run cannot claim.
func main(cur realm) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	tok, led := grc20.NewToken("Run", "RUN", 0, 1, cur)
	key := grc20reg.Register(cross(cur), tok, "")
	factory := chain.PackageAddress("gno.land/r/nt/vesting/v0")
	must(led.Mint(alice, 10))
	must(led.Approve(alice, factory, 10))

	asRun(t0, alice)
	expectPanic("direct user call", func() {
		vesting.CreateCoin(cross(cur), alice, "ugnot", 1, t0, 0, 1, 0, 1, true, "")
	})

	as(t0, alice)
	id := vesting.CreateGRC20(cross(cur), alice, key, 10, t0, 0, 10, 0, 1, false, "run")

	asRun(t0+10, bob)
	expectPanic("not the beneficiary", func() { vesting.Claim(cross(cur), id) })

	asRun(t0+10, alice)
	vesting.Claim(cross(cur), id)
	expectPanic("nothing to release", func() { vesting.Claim(cross(cur), id) })
	if vesting.ClaimedOf(id) != 10 {
		panic("claimed")
	}
	if tok.BalanceOf(alice) != 10 {
		panic("balance")
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

func asRun(unix int64, user address) {
	ctx := testing.GetContext()
	ctx.Time = time.Unix(unix, 0)
	ctx.OriginCaller = user
	ctx.CurrentRealm = runtime.NewRealm(user, "gno.land/e/"+user.String()+"/run")
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
