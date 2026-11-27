// PKGPATH: gno.land/r/nt/multisig/v0/filetests/z_grc20

package z_grc20

import (
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/grc20/v0"
	"gno.land/p/nt/testutils/v0"

	"gno.land/r/nt/grc20reg/v0"
	"gno.land/r/nt/multisig/v0"
)

const t0 int64 = 1_700_000_000

func main(cur realm) {
	setTime(t0)
	tok, led := grc20.NewToken("File Token", "MFT", 0, 1, cur)
	key := grc20reg.Register(cross(cur), tok, "")

	alice := testutils.TestAddress("alice")
	dest := testutils.TestAddress("dest")
	as(t0, alice)
	id := multisig.Create(cross(cur), "grc", alice.String(), 1, 0)
	if err := led.Mint(multisig.SafeAddress(id), 7); err != nil {
		panic(err)
	}
	prop := multisig.ProposeGRC20(cross(cur), id, dest, key, 7, t0+1000, "")
	multisig.Approve(cross(cur), id, prop)
	multisig.Execute(cross(cur), id, prop)
	expectPanic("insufficient balance", func() {
		again := multisig.ProposeGRC20(cross(cur), id, dest, key, 1, t0+1000, "")
		multisig.Approve(cross(cur), id, again)
		multisig.Execute(cross(cur), id, again)
	})

	if tok.BalanceOf(dest) != 7 || tok.BalanceOf(multisig.SafeAddress(id)) != 0 {
		panic("grc20 balances")
	}
	if multisig.ProposalStatus(id, prop) != "executed" {
		panic("status")
	}
	println("ok")
}

func setTime(unix int64) {
	ctx := testing.GetContext()
	ctx.Time = time.Unix(unix, 0)
	testing.SetContext(ctx)
}

func as(unix int64, user address) {
	ctx := testing.GetContext()
	ctx.Time = time.Unix(unix, 0)
	ctx.CurrentRealm = testing.NewUserRealm(user)
	ctx.OriginCaller = user
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
