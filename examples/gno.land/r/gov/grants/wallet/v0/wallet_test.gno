package wallet

import (
	"chain"
	"chain/banker"
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

// Named addresses, so a test relying on the default caller fails instead of
// passing by coincidence. A call that aborts resets the caller: tests set it
// again before every call, in the test's own frame: an override only holds
// on the frame that sets it. The grant flow through GovDAO, and freezing, are
// exercised end to end in gno.land/pkg/integration/testdata/govdao_grants.txtar.
var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
)

func TestOpen(cur realm, t *testing.T) {
	for _, bad := range []string{"", "Gnogolf", "a/b/c", "-x", "x-", "a//b", "a b", "_x", strings.Repeat("a", 65), "../x"} {
		testing.SetRealm(testing.NewUserRealm(alice))
		uassert.AbortsContains(t, cur, "account name", func() { Open(cross(cur), bad) }, bad)
	}
	testing.SetRealm(testing.NewUserRealm(alice))
	Open(cross(cur), "golf")
	uassert.Equal(t, alice, Owner("golf"))
	testing.SetRealm(testing.NewUserRealm(bob))
	uassert.AbortsContains(t, cur, "already open", func() { Open(cross(cur), "golf") })
	testing.SetRealm(testing.NewUserRealm(bob))
	Open(cross(cur), "golf/cup")
	uassert.Equal(t, bob, Owner("golf/cup"))
	uassert.Equal(t, address(""), Owner("nobody"))
	uassert.Equal(t, chain.PackageAddress(path+"#golf/cup"), Address("golf/cup"))
	uassert.PanicsContains(t, cur, "no account can be named", func() { Address("Bad Name") })
}

// Only the owner acts for an account, and the owner can hand it over to an
// address that accepts it.
func TestOnlyTheOwnerActs(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	Open(cross(cur), "owned")
	calls := map[string]func(){
		"Apply":    func() { Apply(cross(cur), "owned", "t", "p", "a:1", 0) },
		"Report":   func() { Report(cross(cur), "owned", 1, "r", "") },
		"Withdraw": func() { Withdraw(cross(cur), "owned", 1) },
		"Spend":    func() { Spend(cross(cur), "owned", bob, "1", "m") },
		"Transfer": func() { TransferOwnership(cross(cur), "owned", bob) },
	}
	for name, call := range calls {
		testing.SetRealm(testing.NewUserRealm(bob))
		uassert.AbortsContains(t, cur, "only the owner", call, name)
	}
	testing.SetRealm(testing.NewUserRealm(bob))
	uassert.AbortsContains(t, cur, "open it first", func() { Spend(cross(cur), "missing", bob, "1", "m") })

	// An offer to an address that can never act, such as another account's,
	// leaves the account with its owner.
	testing.SetRealm(testing.NewUserRealm(alice))
	uassert.AbortsContains(t, cur, "invalid new owner", func() { TransferOwnership(cross(cur), "owned", "nope") })
	testing.SetRealm(testing.NewUserRealm(alice))
	TransferOwnership(cross(cur), "owned", Address("other"))
	uassert.Equal(t, alice, Owner("owned"))
	testing.SetRealm(testing.NewUserRealm(bob))
	uassert.AbortsContains(t, cur, "not offered to you", func() { AcceptOwnership(cross(cur), "owned") })

	testing.SetRealm(testing.NewUserRealm(alice))
	TransferOwnership(cross(cur), "owned", bob)
	uassert.Equal(t, alice, Owner("owned"))
	uassert.True(t, strings.Contains(Render("owned"), "Handover offered to"))
	testing.SetRealm(testing.NewUserRealm(alice))
	uassert.AbortsContains(t, cur, "not offered to you", func() { AcceptOwnership(cross(cur), "owned") })
	testing.SetRealm(testing.NewUserRealm(bob))
	AcceptOwnership(cross(cur), "owned")
	uassert.Equal(t, bob, Owner("owned"))
	testing.SetRealm(testing.NewUserRealm(bob))
	uassert.AbortsContains(t, cur, "not offered to you", func() { AcceptOwnership(cross(cur), "owned") })
	testing.SetRealm(testing.NewUserRealm(alice))
	uassert.AbortsContains(t, cur, "only the owner", func() { TransferOwnership(cross(cur), "owned", alice) })
	page := Render("owned")
	uassert.True(t, strings.Contains(page, "new owner"))
	uassert.False(t, strings.Contains(page, "Handover offered to"))
}

// A spend leaves the account's own address, for exactly its amount, and is
// journaled with its memo.
func TestSpend(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	Open(cross(cur), "spender")
	testing.IssueCoins(Address("spender"), chain.NewCoins(chain.NewCoin("ugnot", 3_000_000)))
	rb := banker.NewReadonlyBanker()
	before := rb.GetCoin(bob, "ugnot")

	for _, bad := range []string{"0", "-1", "1.1234567", "abc", "", "1e6", "1,5"} {
		testing.SetRealm(testing.NewUserRealm(alice))
		uassert.AbortsContains(t, cur, "amount is in GNOT", func() { Spend(cross(cur), "spender", bob, bad, "m") }, bad)
	}
	testing.SetRealm(testing.NewUserRealm(alice))
	uassert.AbortsContains(t, cur, "memo", func() { Spend(cross(cur), "spender", bob, "1", " ") })
	testing.SetRealm(testing.NewUserRealm(alice))
	uassert.AbortsContains(t, cur, "memo", func() { Spend(cross(cur), "spender", bob, "1", "\u200b\u202e") })
	testing.SetRealm(testing.NewUserRealm(alice))
	uassert.AbortsContains(t, cur, "invalid recipient", func() { Spend(cross(cur), "spender", "nope", "1", "m") })
	for _, stuck := range []address{Address("spender"), chain.PackageAddress(path)} {
		testing.SetRealm(testing.NewUserRealm(alice))
		uassert.AbortsContains(t, cur, "invalid recipient", func() { Spend(cross(cur), "spender", stuck, "1", "m") })
	}

	testing.SetRealm(testing.NewUserRealm(alice))
	Spend(cross(cur), "spender", bob, "1.5", "designer, week 1")
	uassert.Equal(t, before+1_500_000, rb.GetCoin(bob, "ugnot"))
	uassert.Equal(t, int64(1_500_000), rb.GetCoin(Address("spender"), "ugnot"))
	page := Render("spender")
	uassert.True(t, strings.Contains(page, "designer, week 1"))
	uassert.True(t, strings.Contains(page, "1.5 GNOT"))
}

func TestParseGNOT(t *testing.T) {
	for in, want := range map[string]int64{"1": 1_000_000, "0.5": 500_000, "12.000001": 12_000_001, " 3 ": 3_000_000, "1.": 1_000_000} {
		got, ok := parseGNOT(in)
		uassert.True(t, ok, in)
		uassert.Equal(t, want, got, in)
	}
	for _, in := range []string{"", ".5", "0", "0.0000001", "+1", "1e3", "9999999999999"} {
		_, ok := parseGNOT(in)
		uassert.False(t, ok, in)
	}
}

// Pages: every route, forms that parse in gnoweb (one per action, each a
// function this realm exports, the account name fixed, textareas
// self-closing) and pagination of the journal.
func TestPages(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	Open(cross(cur), "pages")
	for i := 0; i < pageSize; i++ {
		testing.SetRealm(testing.NewUserRealm(alice))
		TransferOwnership(cross(cur), "pages", alice)
	}
	page := Render("pages")
	for _, fn := range []string{"Apply", "Report", "Spend", "Withdraw", "TransferOwnership", "AcceptOwnership"} {
		uassert.True(t, strings.Contains(page, `<gno-form exec="`+fn+`">`), fn)
	}
	uassert.Equal(t, 6, strings.Count(page, `<gno-input name="name" type="text" value="pages" readonly="true" />`))
	uassert.False(t, strings.Contains(page, "</gno-textarea>"), "textarea with a closing tag")
	uassert.Equal(t, strings.Count(page, "<gno-form "), strings.Count(page, "</gno-form>"))
	uassert.Equal(t, strings.Count(page, "<gno-details "), strings.Count(page, "</gno-details>"))
	uassert.True(t, strings.Contains(page, `href="/r/gov/grants/wallet/v0:pages/_2"`), "older journal")
	uassert.True(t, strings.Contains(Render("pages/_2"), "opened"))
	uassert.False(t, strings.Contains(page, "| opened |"), "the opening is on page 2")

	home := Render("")
	uassert.True(t, strings.Contains(home, `<gno-form exec="Open">`))
	uassert.True(t, strings.Contains(home, "[`pages`](/r/gov/grants/wallet/v0:pages)"))
	for _, p := range []string{"missing", "pages/_0", "pages/_x", "_", "pages/extra", "../x"} {
		uassert.True(t, strings.HasPrefix(Render(p), "# Not found"), p)
	}
	uassert.False(t, strings.HasPrefix(Render("_1"), "# Not found"))
}
