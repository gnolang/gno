// PKGPATH: gno.land/r/nt/vesting/v0/filetests/z_render

package z_render

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
	tok, led := grc20.NewToken("Shown", "SHWN", 0, 1, cur)
	key := grc20reg.Register(cross(cur), tok, "")
	factory := chain.PackageAddress("gno.land/r/nt/vesting/v0")
	must(led.Mint(alice, 4))
	must(led.Approve(alice, factory, 4))

	as(t0, alice)
	id := vesting.CreateGRC20(cross(cur), alice, key, 4, t0, 2, 4, 0, 1, false, "hi [click](http://evil)")
	page := vesting.Render("")
	if !strings.Contains(page, "Vesting vault") {
		panic("home")
	}
	one := vesting.Render(itoa(id))
	if strings.Contains(one, "](http://evil)") {
		panic("link injected: " + one)
	}
	if !strings.Contains(one, "SHWN") {
		panic("symbol")
	}
	if !strings.Contains(one, "irrevocable") {
		panic("irrevocable line")
	}
	if vesting.Render("nope") != "# Vesting vault\n\nUnknown grant.\n" {
		panic("bad path")
	}
	println("ok")
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
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

// Output:
// ok
