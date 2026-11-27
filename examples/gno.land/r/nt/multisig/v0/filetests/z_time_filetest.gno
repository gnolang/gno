// PKGPATH: gno.land/r/nt/multisig/v0/filetests/z_time

package z_time

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
	dest := testutils.TestAddress("dest")
	as(t0, alice)
	id := multisig.Create(cross(cur), "timed", alice.String(), 1, 100)
	locked := multisig.ProposeCoin(cross(cur), id, dest, "ugnot", 1, t0+1000, "")
	multisig.Approve(cross(cur), id, locked)
	expectPanic("timelock", func() { multisig.Execute(cross(cur), id, locked) })
	// Expiry has to sit past the timelock, or the proposal is rejected
	// at creation. This one is ready at t0+100 and dead at t0+150.
	dying := multisig.ProposeCoin(cross(cur), id, dest, "ugnot", 1, t0+150, "")
	multisig.Approve(cross(cur), id, dying)

	as(t0+99, alice)
	expectPanic("timelock", func() { multisig.Execute(cross(cur), id, locked) })

	as(t0+150, alice)
	expectPanic("expired", func() { multisig.Execute(cross(cur), id, dying) })
	testing.IssueCoins(multisig.SafeAddress(id), chain.NewCoins(chain.NewCoin("ugnot", 1)))
	multisig.Execute(cross(cur), id, locked)
	if multisig.ProposalStatus(id, locked) != "executed" {
		panic("locked proposal should execute after the timelock")
	}
	if multisig.ProposalStatus(id, dying) != "expired" {
		panic("dying proposal should be expired")
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
