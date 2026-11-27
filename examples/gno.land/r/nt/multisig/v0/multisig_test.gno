package multisig

import (
	"chain"
	"chain/banker"
	"strconv"
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/grc20/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"

	"gno.land/r/nt/grc20reg/v0"
)

const t0 int64 = 1_700_000_000

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
	carol = testutils.TestAddress("carol")
	dave  = testutils.TestAddress("dave")
	dest  = testutils.TestAddress("dest")
)

func at(unix int64, user address) {
	ctx := testing.GetContext()
	ctx.Time = time.Unix(unix, 0)
	ctx.CurrentRealm = testing.NewUserRealm(user)
	ctx.OriginCaller = user
	testing.SetContext(ctx)
}

func setTime(unix int64) {
	ctx := testing.GetContext()
	ctx.Time = time.Unix(unix, 0)
	testing.SetContext(ctx)
}

func mustPanic(t *testing.T, substr string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected panic containing %q", substr)
			return
		}
		msg := panicText(r)
		if !strings.Contains(msg, substr) {
			t.Errorf("panic %q does not contain %q", msg, substr)
		}
	}()
	fn()
}

func panicText(r any) string {
	switch v := r.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		return "unknown panic"
	}
}

func TestCoinTransferAndExecuteOnce(cur realm, t *testing.T) {
	at(t0, alice)
	var id, prop int64
	before := banker.NewReadonlyBanker().GetCoin(dest, "ugnot")
	func(cur realm) {
		signers := alice.String() + "," + bob.String() + "," + carol.String()
		id = Create(cur, "ops", signers, 2, 0)
		want := chain.DerivePkgSubAddr(cur.PkgPath(), "s/"+strconv.FormatInt(id, 10))
		uassert.Equal(t, want, SafeAddress(id))
		prop = ProposeCoin(cur, id, dest, "ugnot", 40, t0+1000, "payroll")
	}(cross(cur))

	at(t0, dave)
	func(cur realm) {
		mustPanic(t, "not a signer", func() { Approve(cur, id, prop) })
	}(cross(cur))

	at(t0, alice)
	func(cur realm) {
		Approve(cur, id, prop)
		mustPanic(t, "duplicate approval", func() { Approve(cur, id, prop) })
		mustPanic(t, "threshold not met", func() { Execute(cur, id, prop) })
	}(cross(cur))

	at(t0, bob)
	func(cur realm) {
		Approve(cur, id, prop)
		mustPanic(t, "insufficient balance", func() { Execute(cur, id, prop) })
		testing.IssueCoins(SafeAddress(id), chain.NewCoins(chain.NewCoin("ugnot", 40)))
		Execute(cur, id, prop)
		mustPanic(t, "not pending", func() { Execute(cur, id, prop) })
	}(cross(cur))

	uassert.Equal(t, int64(0), Balance(id, "ugnot"))
	uassert.Equal(t, before+40, banker.NewReadonlyBanker().GetCoin(dest, "ugnot"))
	uassert.Equal(t, int64(1), SafeNonce(id))
	uassert.Equal(t, "executed", ProposalStatus(id, prop))
}

func TestTimelockAndExpiry(cur realm, t *testing.T) {
	at(t0, alice)
	var id, locked, dying int64
	func(cur realm) {
		id = Create(cur, "timed", alice.String(), 1, 100)
		locked = ProposeCoin(cur, id, dest, "ugnot", 1, t0+1000, "")
		Approve(cur, id, locked)
		mustPanic(t, "timelock", func() { Execute(cur, id, locked) })
		dying = ProposeCoin(cur, id, dest, "ugnot", 1, t0+150, "")
		Approve(cur, id, dying)
	}(cross(cur))

	at(t0+99, alice)
	func(cur realm) {
		mustPanic(t, "timelock", func() { Execute(cur, id, locked) })
	}(cross(cur))

	at(t0+150, alice)
	func(cur realm) {
		mustPanic(t, "expired", func() { Execute(cur, id, dying) })
	}(cross(cur))

	at(t0+150, alice)
	func(cur realm) {
		testing.IssueCoins(SafeAddress(id), chain.NewCoins(chain.NewCoin("ugnot", 1)))
		Execute(cur, id, locked)
	}(cross(cur))
	uassert.Equal(t, "executed", ProposalStatus(id, locked))
}

func TestRemovalDropsVote(cur realm, t *testing.T) {
	at(t0, alice)
	var id, spend, remove int64
	func(cur realm) {
		signers := alice.String() + "," + bob.String() + "," + carol.String()
		id = Create(cur, "crew", signers, 2, 0)
		spend = ProposeCoin(cur, id, dest, "ugnot", 1, t0+1000, "")
		remove = ProposeRemoveSigner(cur, id, carol, t0+1000, "")
		Approve(cur, id, spend)
		Approve(cur, id, remove)
	}(cross(cur))
	at(t0, carol)
	func(cur realm) {
		Approve(cur, id, spend)
	}(cross(cur))
	at(t0, bob)
	func(cur realm) {
		Approve(cur, id, remove)
		Execute(cur, id, remove)
		mustPanic(t, "threshold not met", func() { Execute(cur, id, spend) })
	}(cross(cur))
	uassert.False(t, IsSigner(id, carol))
	uassert.Equal(t, 1, ApprovalCount(id, spend))
}

func TestGRC20AndRender(cur realm, t *testing.T) {
	setTime(t0)
	tok, led := grc20.NewToken("Multisig Token", "MSIG", 0, 1, cur)
	key := grc20reg.Register(cross(cur), tok, "")

	at(t0, alice)
	var id, prop int64
	func(cur realm) {
		id = Create(cur, "token-safe", alice.String(), 1, 0)
		if err := led.Mint(SafeAddress(id), 25); err != nil {
			panic(err)
		}
		WatchToken(cur, id, key)
		prop = ProposeGRC20(cur, id, dest, key, 25, t0+1000, "grant")
		Approve(cur, id, prop)
		Execute(cur, id, prop)
		mustPanic(t, "insufficient balance", func() {
			again := ProposeGRC20(cur, id, dest, key, 1, t0+1000, "")
			Approve(cur, id, again)
			Execute(cur, id, again)
		})
	}(cross(cur))

	uassert.Equal(t, int64(25), tok.BalanceOf(dest))
	uassert.Equal(t, int64(0), tok.BalanceOf(SafeAddress(id)))
	page := Render(strconv.FormatInt(id, 10))
	if !strings.Contains(page, "MSIG") || !strings.Contains(page, "Timelock") {
		t.Errorf("safe page missing fields:\n%s", page)
	}
	if !strings.Contains(Render(""), "Multisig treasury") {
		t.Error("home page missing title")
	}
	propPage := Render(strconv.FormatInt(id, 10) + "/" + strconv.FormatInt(prop, 10))
	if !strings.Contains(propPage, "grant") {
		t.Errorf("proposal page missing memo:\n%s", propPage)
	}
}

func TestDeposit(cur realm, t *testing.T) {
	at(t0, alice)
	var id int64
	func(cur realm) {
		id = Create(cur, "deposit", alice.String(), 1, 0)
	}(cross(cur))
	// -send credits the realm before the call. The test banker needs
	// both the balance and the origin-send allowance.
	testing.IssueCoins(chain.PackageAddress(realmPath), chain.NewCoins(chain.NewCoin("ugnot", 50)))
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", 50)))
	at(t0, alice)
	func(cur realm) {
		Deposit(cur, id, "ugnot", 50)
	}(cross(cur))
	uassert.Equal(t, int64(50), Balance(id, "ugnot"))
}
