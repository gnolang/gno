// PKGPATH: gno.land/r/nt/multisig/v0/filetests/z_coin

package z_coin

import (
	"chain"
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/testutils/v0"

	"gno.land/r/nt/multisig/v0"
)

const t0 int64 = 1_700_000_000

func main(cur realm) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	dave := testutils.TestAddress("dave")
	dest := testutils.TestAddress("dest")
	signers := alice.String() + "," + bob.String()

	as(t0, alice)
	id := multisig.Create(cross(cur), "coin", signers, 2, 0)
	prop := multisig.ProposeCoin(cross(cur), id, dest, "ugnot", 40, t0+1000, "pay")
	multisig.Approve(cross(cur), id, prop)
	expectPanic("duplicate approval", func() { multisig.Approve(cross(cur), id, prop) })
	expectPanic("threshold not met", func() { multisig.Execute(cross(cur), id, prop) })

	as(t0, dave)
	expectPanic("not a signer", func() { multisig.Approve(cross(cur), id, prop) })
	expectPanic("not a signer", func() { multisig.Execute(cross(cur), id, prop) })

	as(t0, bob)
	multisig.Approve(cross(cur), id, prop)
	expectPanic("insufficient balance", func() { multisig.Execute(cross(cur), id, prop) })
	testing.IssueCoins(multisig.SafeAddress(id), chain.NewCoins(chain.NewCoin("ugnot", 40)))
	multisig.Execute(cross(cur), id, prop)
	expectPanic("not pending", func() { multisig.Execute(cross(cur), id, prop) })

	if multisig.Balance(id, "ugnot") != 0 {
		panic("safe should be empty")
	}
	if multisig.SafeNonce(id) != 1 {
		panic("nonce")
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
