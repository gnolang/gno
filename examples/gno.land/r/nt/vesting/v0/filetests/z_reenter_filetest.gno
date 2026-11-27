// PKGPATH: gno.land/r/nt/vesting/v0/filetests/z_reenter

package z_reenter

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

var attacks int

// Attack is what a token callback would call. The canonical GRC20
// teller does not call out. Calling Attack by hand cannot spend a
// grant this package does not own.
func Attack(cur realm, id int64) {
	attacks++
	vesting.Claim(cross(cur), id)
}

func main(cur realm) {
	alice := testutils.TestAddress("alice")
	tok, led := grc20.NewToken("Re", "RE", 0, 1, cur)
	key := grc20reg.Register(cross(cur), tok, "")
	factory := chain.PackageAddress("gno.land/r/nt/vesting/v0")
	must(led.Mint(alice, 8))
	must(led.Approve(alice, factory, 8))

	as(t0, alice)
	id := vesting.CreateGRC20(cross(cur), alice, key, 8, t0, 0, 4, 0, 1, true, "re")
	as(t0+4, alice)
	vesting.Claim(cross(cur), id)
	if attacks != 0 {
		panic("claim invoked a callback")
	}

	expectPanic("not the beneficiary", func() { Attack(cross(cur), id) })
	if attacks != 1 {
		panic("attack counter")
	}
	if vesting.ClaimedOf(id) != 8 {
		panic("books moved")
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
