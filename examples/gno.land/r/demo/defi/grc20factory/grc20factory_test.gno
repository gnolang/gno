package grc20factory

import (
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/ufmt/v0"
)

func TestReadOnlyPublicMethods(cur realm, t *testing.T) {
	admin := testutils.TestAddress("admin")
	bob := testutils.TestAddress("bob")
	carl := testutils.TestAddress("carl")

	type test struct {
		name    string
		balance uint256
		fn      func() uint256
	}

	checkBalances := func(step string, totSup, balAdm, balBob, allowAdmBob, balCarl uint256) {
		tests := []test{
			{"TotalSupply", totSup, func() uint256 { return TotalSupply("FOO") }},
			{"BalanceOf(admin)", balAdm, func() uint256 { return BalanceOf("FOO", admin) }},
			{"BalanceOf(bob)", balBob, func() uint256 { return BalanceOf("FOO", bob) }},
			{"Allowance(admin, bob)", allowAdmBob, func() uint256 { return Allowance("FOO", admin, bob) }},
			{"BalanceOf(carl)", balCarl, func() uint256 { return BalanceOf("FOO", carl) }},
		}
		for _, tc := range tests {
			reason := ufmt.Sprintf("%s.%s - %s", step, tc.name, "balances do not match")
			uassert.Equal(t, tc.balance, tc.fn(), reason)
		}
	}

	// admin creates FOO and BAR.
	testing.SetOriginCaller(admin)
	NewWithAdmin(cross(cur), "Foo", "FOO", 3, 1_111_111_000, 5_555, admin)
	NewWithAdmin(cross(cur), "Bar", "BAR", 3, 2_222_000, 6_666, admin)
	uassert.Equal(t, Bank("FOO").ID(), "gno.land/r/demo/defi/grc20factory.FOO.0000001")
	uassert.Equal(t, Bank("BAR").ID(), "gno.land/r/demo/defi/grc20factory.BAR.0000002")
	checkBalances("step1", 1_111_111_000, 1_111_111_000, 0, 0, 0)

	// admin mints to bob.
	mustGetInstance("FOO").ledger.Mint(bob, 333_333_000)
	checkBalances("step2", 1_444_444_000, 1_111_111_000, 333_333_000, 0, 0)

	// carl uses the faucet.
	testing.SetOriginCaller(carl)
	Faucet(cross(cur), "FOO")
	checkBalances("step3", 1_444_449_555, 1_111_111_000, 333_333_000, 0, 5_555)

	// admin gives to bob some allowance.
	testing.SetOriginCaller(admin)
	Approve(cross(cur), "FOO", bob, 1_000_000)
	checkBalances("step4", 1_444_449_555, 1_111_111_000, 333_333_000, 1_000_000, 5_555)

	// bob uses a part of the allowance.
	testing.SetOriginCaller(bob)
	TransferFrom(cross(cur), "FOO", admin, carl, 400_000)
	checkBalances("step5", 1_444_449_555, 1_110_711_000, 333_333_000, 600_000, 405_555)

	// bob uses a part of the allowance.
	testing.SetOriginCaller(bob)
	TransferFrom(cross(cur), "FOO", admin, carl, 600_000)
	checkBalances("step6", 1_444_449_555, 1_110_111_000, 333_333_000, 0, 1_005_555)
}

func TestAbove2Pow63(cur realm, t *testing.T) {
	admin := testutils.TestAddress("admin_big")
	bob := testutils.TestAddress("bob_big")
	big := uint256(1) << 64

	testing.SetOriginCaller(admin)
	NewWithAdmin(cross(cur), "Big", "BIG", 18, big, big, admin)
	uassert.Equal(t, big, TotalSupply("BIG"))

	testing.SetOriginCaller(bob)
	Faucet(cross(cur), "BIG")
	uassert.Equal(t, big, BalanceOf("BIG", bob))
	uassert.Equal(t, 2*big, TotalSupply("BIG"))

	testing.SetOriginCaller(bob)
	Transfer(cross(cur), "BIG", admin, big)
	uassert.Equal(t, 2*big, BalanceOf("BIG", admin))
	uassert.Equal(t, "**"+admin.String()+"** balance: 36893488147419103232\n\n", Render("BIG/balance/"+admin.String()))
}

// Anyone can create a factory token, so its name must not inject markdown
// into the factory's pages.
func TestRenderEscapesTokenName(cur realm, t *testing.T) {
	const evil = "[Claim reward](https://evil.example)"
	testing.SetOriginCaller(testutils.TestAddress("evil_creator"))
	New(cross(cur), evil, "EVILNAME", 0, 0, 0)

	for _, path := range []string{"", "EVILNAME"} {
		out := Render(path)
		uassert.False(t, strings.Contains(out, evil), "live link in Render("+path+"): "+out)
		uassert.True(t, strings.Contains(out, `\[Claim reward\]`), "name not escaped in Render("+path+"): "+out)
	}

	// A path segment reflected into the balance page is escaped too.
	out := Render("EVILNAME/balance/[x](y)")
	uassert.False(t, strings.Contains(out, "[x](y)"), "path segment rendered raw: "+out)
}
