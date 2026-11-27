// PKGPATH: gno.land/r/nt/vesting/v0/filetests/z_coin

package z_coin

import (
	"chain"
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/testutils/v0"

	"gno.land/r/nt/vesting/v0"
)

const t0 int64 = 1_700_000_000

// A foreign realm cannot fund a coin grant. A direct user call is the
// only caller allowed to spend the envelope, and even a user-shaped
// caller fails closed when the envelope was paid to this package
// rather than to the vault.
func main(cur realm) {
	alice := testutils.TestAddress("alice")
	expectPanic("direct user call", func() {
		vesting.CreateCoin(cross(cur), alice, "ugnot", 10, t0, 0, 10, 0, 1, true, "")
	})

	as(t0, alice)
	testing.IssueCoins(chain.PackageAddress("gno.land/r/nt/vesting/v0"), chain.NewCoins(chain.NewCoin("ugnot", 10)))
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", 10)))
	expectPanic("did not receive the origin send", func() {
		vesting.CreateCoin(cross(cur), alice, "ugnot", 10, t0, 0, 10, 0, 1, true, "")
	})
	if vesting.Count() != 0 {
		panic("rejected funding created a grant")
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
