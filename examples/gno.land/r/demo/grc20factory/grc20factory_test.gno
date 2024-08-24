package foo20

import (
	"std"
	"testing"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

func TestReadOnlyPublicMethods(t *testing.T) {
	admin := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	manfred := std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
	unknown := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // valid but never used.
	NewWithAdmin("Foo", "FOO", 4, 10_000*1_000_000, 0, admin)
	NewWithAdmin("Bar", "BAR", 4, 10_000*1_000, 0, admin)
	mustGetInstance("FOO").banker.Mint(manfred, 100_000_000)

	type balanceTest struct {
		name    string
		balance uint64
		fn      func() uint64
	}

	// Check balances (initial state).
	{
		balanceTests := []balanceTest{
			{"TotalSupply", 10_100_000_000, func() uint64 { return TotalSupply("FOO") }},
			{"BalanceOf(admin)", 10_000_000_000, func() uint64 { return BalanceOf("FOO", admin) }},
			{"BalanceOf(manfred)", 100_000_000, func() uint64 { return BalanceOf("FOO", manfred) }},
			{"Allowance(admin, manfred)", 0, func() uint64 { return Allowance("FOO", admin, manfred) }},
			{"BalanceOf(unknown)", 0, func() uint64 { return BalanceOf("FOO", unknown) }},
		}
		for _, tc := range balanceTests {
			urequire.Equal(t, tc.balance, tc.fn(), "balance does not match")
		}
	}

	// unknown uses the faucet.
	std.TestSetOrigCaller(unknown)
	urequire.PanicsWithMessage(t, "faucet disabled for this token", func() {
		Faucet("FOO")
	}, "Expected panic message not received")

	// Check balances after faucet.
	balanceTestsAfterFaucet := []balanceTest{
		{"TotalSupply", 10_100_000_000, func() uint64 { return TotalSupply("FOO") }},
		{"BalanceOf(admin)", 10_000_000_000, func() uint64 { return BalanceOf("FOO", admin) }},
		{"BalanceOf(manfred)", 100_000_000, func() uint64 { return BalanceOf("FOO", manfred) }},
		{"Allowance(admin, manfred)", 0, func() uint64 { return Allowance("FOO", admin, manfred) }},
		{"BalanceOf(unknown)", 0, func() uint64 { return BalanceOf("FOO", unknown) }},
	}

	for _, tc := range balanceTestsAfterFaucet {
		t.Run(tc.name, func(t *testing.T) {
			urequire.Equal(t, tc.balance, tc.fn(), "balance does not match after faucet")
		})
	}

	// Check Render
	renderTests := []struct {
		path     string
		expected string
	}{
		{"", "Token: BAR, Admin: " + admin.String() + "\nToken: FOO, Admin: " + admin.String() + "\n"},
		{"FOO", mustGetInstance("FOO").banker.RenderHome()},
		{"FOO/balance/g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq", "100000000"},
		{"unknown/path", "404\n"},
	}
	for _, tc := range renderTests {
		t.Run("Render-"+tc.path, func(t *testing.T) {
			result := Render(tc.path)
			urequire.Equal(t, tc.expected, result, ufmt.Sprintf("render does not match for path %s", tc.path))
		})
	}

	// Change admin
	std.TestSetOrigCaller(admin)
	SetAdmin("FOO", unknown)
	newAdmin := mustGetInstance("FOO").admin.Owner()
	urequire.Equal(t, unknown.String(), newAdmin.String(), "admin change failed")

	// Verify the total number of instances
	expected := uint64(2)
	actual := TotalInstances()
	urequire.Equal(t, expected, actual, ufmt.Sprintf("expected 2 instances but got %d", actual))
}
