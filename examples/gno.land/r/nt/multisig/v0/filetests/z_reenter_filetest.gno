// PKGPATH: gno.land/r/nt/multisig/v0/filetests/z_reenter

package z_reenter

import (
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/testutils/v0"

	"gno.land/r/nt/multisig/v0"
)

const t0 int64 = 1_700_000_000

var attacks int

// Attack is the callback a dynamic call would have made. Execute must
// not enter it. Calling it by hand cannot spend the safe: this package
// is not a signer.
func Attack(cur realm, safeID, propID int64) {
	attacks++
	multisig.Execute(cross(cur), safeID, propID)
}

func main(cur realm) {
	me := cur.PkgPath()
	alice := testutils.TestAddress("alice")
	as(t0, alice)
	id := multisig.Create(cross(cur), "call", alice.String(), 1, 0)
	prop := multisig.ProposeCall(cross(cur), id, me, "payload-1", t0+1000, "")
	multisig.Approve(cross(cur), id, prop)
	multisig.Execute(cross(cur), id, prop)

	if attacks != 0 {
		panic("execute invoked the target")
	}
	if multisig.SafeNonce(id) != 1 {
		panic("nonce")
	}

	ctx := testing.GetContext()
	ctx.Time = time.Unix(t0, 0)
	ctx.CurrentRealm = testing.NewCodeRealm(me)
	testing.SetContext(ctx)
	payload := multisig.Consume(cross(cur), id, prop)
	if payload != "payload-1" {
		panic("payload")
	}
	expectPanic("already consumed", func() { multisig.Consume(cross(cur), id, prop) })

	expectPanic("not a signer", func() { Attack(cross(cur), id, prop) })
	if attacks != 1 {
		panic("attack counter")
	}
	if multisig.SafeNonce(id) != 1 {
		panic("reenter moved the nonce")
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
