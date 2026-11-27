package vesting

import (
	"chain"
	"chain/banker"
	"strconv"
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/grc20/v0"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"

	"gno.land/r/nt/grc20reg/v0"
)

const t0 int64 = 1_700_000_000

var nextTok seqid.ID

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

func factory() address { return chain.PackageAddress("gno.land/r/nt/vesting/v0") }

// allow sets this message's origin-send envelope to exactly one coin
// and clears what the test banker has already counted as spent. Each
// CreateCoin requires that exact envelope.
func allow(denom string, amount int64) {
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin(denom, amount)))
	testing.SetOriginSpend(chain.Coins{})
}

func credit(denom string, amount int64) {
	testing.IssueCoins(factory(), chain.NewCoins(chain.NewCoin(denom, amount)))
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

func TestCoinClaimRevokeAndIsolation(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carol := testutils.TestAddress("carol")
	at(t0, alice)
	var idA, idB int64
	func(cur realm) {
		allow("ugnot", 1000)
		credit("ugnot", 1000)
		idA = CreateCoin(cur, alice, "ugnot", 1000, t0, 100, 1000, 0, 1, true, "alice")
		uassert.Equal(t, chain.DerivePkgSubAddr(cur.PkgPath(), "g/"+strconv.FormatInt(idA, 10)), GrantAddress(idA))
		allow("ugnot", 50)
		credit("ugnot", 50)
		idB = CreateCoin(cur, bob, "ugnot", 50, t0, 0, 100, 0, 1, false, "bob")
	}(cross(cur))

	uassert.True(t, idA != idB)
	uassert.Equal(t, int64(1000), BalanceOf(idA))
	uassert.Equal(t, int64(50), BalanceOf(idB))

	at(t0, bob)
	func(cur realm) {
		mustPanic(t, "not the beneficiary", func() { Claim(cur, idA) })
		mustPanic(t, "nothing to release", func() { Claim(cur, idB) })
	}(cross(cur))

	setTime(t0 + 100)
	at(t0+100, alice)
	beforeAlice := banker.NewReadonlyBanker().GetCoin(alice, "ugnot")
	beforeBob := banker.NewReadonlyBanker().GetCoin(bob, "ugnot")
	func(cur realm) {
		uassert.Equal(t, int64(100), VestedOf(idA))
		uassert.Equal(t, int64(100), ClaimableOf(idA))
		Claim(cur, idA)
		mustPanic(t, "nothing to release", func() { Claim(cur, idA) })
		mustPanic(t, "irrevocable", func() { Revoke(cur, idB) })
	}(cross(cur))
	uassert.Equal(t, int64(100), ClaimedOf(idA))
	uassert.Equal(t, int64(900), BalanceOf(idA))
	uassert.Equal(t, int64(50), BalanceOf(idB))
	uassert.Equal(t, beforeBob, banker.NewReadonlyBanker().GetCoin(bob, "ugnot"))
	uassert.Equal(t, beforeAlice+100, banker.NewReadonlyBanker().GetCoin(alice, "ugnot"))

	at(t0+100, alice)
	func(cur realm) {
		Revoke(cur, idA)
		mustPanic(t, "already revoked", func() { Revoke(cur, idA) })
	}(cross(cur))
	uassert.True(t, RevokedOf(idA))
	uassert.Equal(t, int64(900), ReturnedOf(idA))
	uassert.Equal(t, int64(0), BalanceOf(idA))
	uassert.Equal(t, beforeAlice+1000, banker.NewReadonlyBanker().GetCoin(alice, "ugnot"))

	// bob's grant is untouched and irrevocable. He is not the admin.
	setTime(t0 + 100)
	at(t0+100, bob)
	func(cur realm) {
		mustPanic(t, "not the admin", func() { Revoke(cur, idB) })
		Claim(cur, idB)
		mustPanic(t, "nothing to release", func() { Claim(cur, idB) })
	}(cross(cur))
	uassert.Equal(t, int64(50), ClaimedOf(idB))
	uassert.Equal(t, int64(0), OutstandingOf(idB))
	uassert.False(t, RevocableOf(idB))
	uassert.Equal(t, int64(0), BalanceOf(idB))
	uassert.Equal(t, int64(0), BalanceOf(idA))

	at(t0+100, carol)
	func(cur realm) {
		mustPanic(t, "unknown grant", func() { Claim(cur, 0) })
		mustPanic(t, "unknown grant", func() { Claim(cur, 999_999) })
	}(cross(cur))
}

func TestBeneficiaryAndAdmin(cur realm, t *testing.T) {
	alice := testutils.TestAddress("aliceb")
	bob := testutils.TestAddress("bobb")
	cara := testutils.TestAddress("carab")
	at(t0, alice)
	var id int64
	func(cur realm) {
		allow("ugnot", 100)
		credit("ugnot", 100)
		id = CreateCoin(cur, bob, "ugnot", 100, t0, 0, 10, 0, 1, true, "move")
		mustPanic(t, "not the beneficiary", func() { SetBeneficiary(cur, id, cara) })
	}(cross(cur))
	at(t0, bob)
	func(cur realm) {
		mustPanic(t, "unchanged", func() { SetBeneficiary(cur, id, bob) })
		SetBeneficiary(cur, id, cara)
		mustPanic(t, "not the beneficiary", func() { Claim(cur, id) })
	}(cross(cur))
	setTime(t0 + 10)
	at(t0+10, cara)
	func(cur realm) {
		Claim(cur, id)
	}(cross(cur))
	uassert.Equal(t, int64(100), banker.NewReadonlyBanker().GetCoin(cara, "ugnot"))
	uassert.Equal(t, cara, BeneficiaryOf(id))

	at(t0+10, alice)
	func(cur realm) {
		TransferAdmin(cur, id, bob)
		mustPanic(t, "not the admin", func() { Revoke(cur, id) })
	}(cross(cur))
	uassert.Equal(t, bob, AdminOf(id))
	at(t0+10, bob)
	func(cur realm) {
		Revoke(cur, id)
	}(cross(cur))
	// A fully vested revoke returns nothing and still closes the grant.
	uassert.True(t, RevokedOf(id))
	uassert.Equal(t, int64(0), ReturnedOf(id))
}

func TestPaymentRejection(cur realm, t *testing.T) {
	alice := testutils.TestAddress("pay")
	at(t0, alice)
	func(cur realm) {
		testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", 5)))
		mustPanic(t, "payment must match", func() {
			CreateCoin(cur, alice, "ugnot", 9, t0, 0, 1, 0, 1, true, "")
		})
		testing.SetOriginSend(chain.NewCoins(chain.NewCoin("atom", 9)))
		mustPanic(t, "payment must match", func() {
			CreateCoin(cur, alice, "ugnot", 9, t0, 0, 1, 0, 1, true, "")
		})
		testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", 9), chain.NewCoin("atom", 1)))
		mustPanic(t, "payment must match", func() {
			CreateCoin(cur, alice, "ugnot", 9, t0, 0, 1, 0, 1, true, "")
		})
		testing.SetOriginSend(chain.Coins{})
		mustPanic(t, "payment must match", func() {
			CreateCoin(cur, alice, "ugnot", 9, t0, 0, 1, 0, 1, true, "")
		})
		allow("ugnot", 9)
		credit("ugnot", 9)
		mustPanic(t, "cliff", func() {
			CreateCoin(cur, alice, "ugnot", 9, t0, 5, 4, 0, 1, true, "")
		})
		mustPanic(t, "unknown curve", func() {
			CreateCoin(cur, alice, "ugnot", 9, t0, 0, 4, 0, 3, true, "")
		})
	}(cross(cur))
	// The rejected calls must not leave the guard stuck.
	at(t0, alice)
	func(cur realm) {
		allow("ugnot", 3)
		credit("ugnot", 3)
		id := CreateCoin(cur, alice, "ugnot", 3, t0, 0, 1, 0, 1, true, "after")
		uassert.Equal(t, "after", LabelOf(id))
	}(cross(cur))
}

func TestGRC20AndUnregistered(cur realm, t *testing.T) {
	alice := testutils.TestAddress("grc")
	bob := testutils.TestAddress("grcb")
	at(t0, alice)
	var key string
	var id int64
	func(cur realm) {
		sym := "T" + strconv.FormatInt(int64(nextTok.Next()), 10)
		tok, led := grc20.NewToken("Test", sym, 0, seqid.ID(0), cur)
		key = grc20reg.Register(cross(cur), tok, "")
		uassert.NoError(t, led.Mint(alice, 200))
		uassert.NoError(t, led.Approve(alice, factory(), 80))
		mustPanic(t, "invalid token", func() {
			CreateGRC20(cur, bob, "missing.token", 10, t0, 0, 10, 0, 1, true, "")
		})
		mustPanic(t, "insufficient allowance", func() {
			CreateGRC20(cur, bob, key, 81, t0, 0, 10, 0, 1, true, "")
		})
		id = CreateGRC20(cur, bob, key, 80, t0, 0, 10, 5, 2, true, "token")
	}(cross(cur))
	uassert.Equal(t, int64(80), BalanceOf(id))
	uassert.True(t, strings.Contains(Render(strconv.FormatInt(id, 10)), "Period: 5"))
	setTime(t0 + 5)
	at(t0+5, bob)
	func(cur realm) {
		Claim(cur, id)
	}(cross(cur))
	uassert.Equal(t, int64(40), ClaimedOf(id))
	uassert.Equal(t, int64(40), BalanceOf(id))
	tok := grc20reg.Get(key)
	uassert.Equal(t, int64(40), tok.BalanceOf(bob))
	at(t0+5, alice)
	func(cur realm) {
		Revoke(cur, id)
	}(cross(cur))
	uassert.Equal(t, int64(40), ReturnedOf(id))
	uassert.Equal(t, int64(160), tok.BalanceOf(alice))
	uassert.Equal(t, int64(0), BalanceOf(id))
	uassert.Equal(t, int64(0), ClaimableOf(id))
}

func TestDrainedGrantDoesNotChangeBooks(cur realm, t *testing.T) {
	alice := testutils.TestAddress("drain")
	at(t0+5, alice)
	var id int64
	func(cur realm) {
		allow("ugnot", 20)
		credit("ugnot", 20)
		id = CreateCoin(cur, alice, "ugnot", 20, t0, 0, 10, 0, 1, true, "drain")
		sub := cur.Sub("g/" + strconv.FormatInt(id, 10))
		b := banker.NewBanker(banker.BankerTypeRealmSend, sub)
		b.SendCoins(sub.Address(), alice, chain.NewCoins(chain.NewCoin("ugnot", 20)))
		mustPanic(t, "insufficient balance", func() { Claim(cur, id) })
		mustPanic(t, "insufficient balance", func() { Revoke(cur, id) })
	}(cross(cur))
	uassert.Equal(t, int64(0), ClaimedOf(id))
	uassert.False(t, RevokedOf(id))
	uassert.Equal(t, int64(10), ClaimableOf(id))
}

func TestReentrancyGuard(cur realm, t *testing.T) {
	alice := testutils.TestAddress("re")
	busy = true
	at(t0, alice)
	func(cur realm) {
		mustPanic(t, "reentrant", func() { Claim(cur, 1) })
		mustPanic(t, "reentrant", func() { Revoke(cur, 1) })
		mustPanic(t, "reentrant", func() { SetBeneficiary(cur, 1, alice) })
		mustPanic(t, "reentrant", func() { TransferAdmin(cur, 1, alice) })
		mustPanic(t, "reentrant", func() {
			CreateCoin(cur, alice, "ugnot", 1, t0, 0, 1, 0, 1, true, "")
		})
		mustPanic(t, "reentrant", func() {
			CreateGRC20(cur, alice, "nope", 1, t0, 0, 1, 0, 1, true, "")
		})
	}(cross(cur))
	busy = false
}

func TestDonationDoesNotIncreaseClaim(cur realm, t *testing.T) {
	alice := testutils.TestAddress("don")
	at(t0+10, alice)
	var id int64
	func(cur realm) {
		allow("ugnot", 100)
		credit("ugnot", 100)
		id = CreateCoin(cur, alice, "ugnot", 100, t0, 0, 10, 0, 1, true, "don")
	}(cross(cur))
	testing.IssueCoins(GrantAddress(id), chain.NewCoins(chain.NewCoin("ugnot", 40)))
	uassert.Equal(t, int64(140), BalanceOf(id))
	uassert.Equal(t, int64(100), ClaimableOf(id))
	uassert.Equal(t, int64(100), OutstandingOf(id))
	at(t0+10, alice)
	func(cur realm) {
		Claim(cur, id)
		mustPanic(t, "nothing to release", func() { Claim(cur, id) })
	}(cross(cur))
	uassert.Equal(t, int64(100), ClaimedOf(id))
	uassert.Equal(t, int64(0), OutstandingOf(id))
	uassert.Equal(t, int64(40), BalanceOf(id))
}

func TestRolesStaySeparate(cur realm, t *testing.T) {
	alice := testutils.TestAddress("rolea")
	bob := testutils.TestAddress("roleb")
	at(t0+10, alice)
	var id int64
	func(cur realm) {
		allow("ugnot", 10)
		credit("ugnot", 10)
		id = CreateCoin(cur, bob, "ugnot", 10, t0, 0, 10, 0, 1, true, "roles")
		mustPanic(t, "not the beneficiary", func() { Claim(cur, id) })
	}(cross(cur))
	at(t0+10, bob)
	func(cur realm) {
		mustPanic(t, "not the admin", func() { Revoke(cur, id) })
		mustPanic(t, "not the admin", func() { TransferAdmin(cur, id, alice) })
		Claim(cur, id)
	}(cross(cur))
	uassert.Equal(t, int64(10), ClaimedOf(id))
	uassert.Equal(t, int64(0), BalanceOf(id))
	uassert.Equal(t, int64(0), OutstandingOf(id))
}

func TestRender(cur realm, t *testing.T) {
	alice := testutils.TestAddress("rend")
	nasty := "hello # pwned [click](http://evil)"
	at(t0, alice)
	var id int64
	func(cur realm) {
		allow("ugnot", 8)
		credit("ugnot", 8)
		mustPanic(t, "invalid label", func() {
			CreateCoin(cur, alice, "ugnot", 8, t0, 4, 8, 0, 1, true, "hello\n\n# pwned")
		})
		allow("ugnot", 8)
		id = CreateCoin(cur, alice, "ugnot", 8, t0, 4, 8, 0, 1, true, nasty)
	}(cross(cur))
	home := Render("")
	uassert.True(t, strings.Contains(home, "Vesting vault"))
	page := Render(strconv.FormatInt(id, 10))
	uassert.True(t, strings.Contains(page, "Cliff: 4"))
	uassert.True(t, strings.Contains(page, "nothing vested yet"))
	uassert.False(t, strings.Contains(page, "](http://evil)"))
	uassert.Equal(t, "# Vesting vault\n\nUnknown grant.\n", Render("nope"))
	uassert.Equal(t, "# Vesting vault\n\nUnknown grant.\n", Render("1/2"))
	uassert.Equal(t, "# Vesting vault\n\nUnknown grant.\n", Render(strings.Repeat("1", 65)))
	setTime(-1)
	uassert.Equal(t, "# Vesting vault\n\nUnknown grant.\n", Render(strconv.FormatInt(id, 10)))
	setTime(t0 + 6)
	mid := Render(strconv.FormatInt(id, 10))
	uassert.True(t, strings.Contains(mid, "Status: vesting."))
	setTime(t0 + 8)
	done := Render(strconv.FormatInt(id, 10))
	uassert.True(t, strings.Contains(done, "fully vested"))
	at(t0+8, alice)
	func(cur realm) {
		Revoke(cur, id)
	}(cross(cur))
	rev := Render(strconv.FormatInt(id, 10))
	uassert.True(t, strings.Contains(rev, "Status: revoked"))
	uassert.True(t, strings.Contains(rev, "already revoked"))
	uassert.False(t, strings.Contains(rev, "may still revoke"))
	setTime(t0)
}

func TestRenderLatestPage(cur realm, t *testing.T) {
	alice := testutils.TestAddress("page")
	at(t0, alice)
	before := Count()
	var first, last int64
	func(cur realm) {
		for i := 0; i < 21; i++ {
			allow("ugnot", 1)
			credit("ugnot", 1)
			id := CreateCoin(cur, alice, "ugnot", 1, t0, 0, 1, 0, 1, false, "")
			if i == 0 {
				first = id
			}
			last = id
		}
	}(cross(cur))
	uassert.Equal(t, before+21, Count())
	home := Render("")
	uassert.True(t, strings.Contains(home, "Showing the latest 20."))
	uassert.True(t, strings.Contains(home, ":"+strconv.FormatInt(last, 10)+")"))
	uassert.False(t, strings.Contains(home, ":"+strconv.FormatInt(first, 10)+")"))
	uassert.True(t, strings.Contains(Render(strconv.FormatInt(last, 10)), "Grant "+strconv.FormatInt(last, 10)))
	emptyish := Render("/")
	uassert.True(t, strings.Contains(emptyish, "Grants:"))
}
