package gnoland

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// Vesting balance Verify tests

func TestBalance_VestingVerify(t *testing.T) {
	validAddress := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	nonEmptyAmount := std.NewCoins(std.NewCoin("test", 100))
	vestingAmount := std.NewCoins(std.NewCoin("test", 50))

	tests := []struct {
		name      string
		balance   Balance
		expectErr bool
	}{
		{
			"valid vesting",
			Balance{
				Address: validAddress, Amount: nonEmptyAmount,
				OriginalVesting: vestingAmount, VestingStartTime: 100, VestingEndTime: 200,
			},
			false,
		},
		{
			"vesting start >= end",
			Balance{
				Address: validAddress, Amount: nonEmptyAmount,
				OriginalVesting: vestingAmount, VestingStartTime: 200, VestingEndTime: 100,
			},
			true,
		},
		{
			"vesting exceeds balance",
			Balance{
				Address: validAddress, Amount: vestingAmount,
				OriginalVesting: nonEmptyAmount, VestingStartTime: 100, VestingEndTime: 200,
			},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.balance.Verify()
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Vesting balance Parse tests

func TestBalance_VestingParse(t *testing.T) {
	validAddress := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	tests := []struct {
		name      string
		entry     string
		expected  Balance
		expectErr bool
	}{
		{
			"no vesting (backward compat)",
			"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=100test",
			Balance{Address: validAddress, Amount: std.NewCoins(std.NewCoin("test", 100))},
			false,
		},
		{
			"with vesting",
			"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=100test;vesting=50test;start=100;end=200",
			Balance{
				Address: validAddress, Amount: std.NewCoins(std.NewCoin("test", 100)),
				OriginalVesting:  std.NewCoins(std.NewCoin("test", 50)),
				VestingStartTime: 100, VestingEndTime: 200,
			},
			false,
		},
		{
			"invalid vesting amount",
			"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=100test;vesting=invalid",
			Balance{},
			true,
		},
		{
			"invalid vesting start time",
			"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=100test;vesting=50test;start=abc;end=200",
			Balance{},
			true,
		},
		{
			"unknown vesting option",
			"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=100test;vesting=50test;foo=bar",
			Balance{},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			balance := Balance{}
			err := balance.Parse(tc.entry)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected.Address, balance.Address)
				assert.True(t, tc.expected.Amount.IsEqual(balance.Amount))
				assert.True(t, tc.expected.OriginalVesting.IsEqual(balance.OriginalVesting))
				assert.Equal(t, tc.expected.VestingStartTime, balance.VestingStartTime)
				assert.Equal(t, tc.expected.VestingEndTime, balance.VestingEndTime)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Vesting balance String / MarshalAmino tests

func TestBalance_VestingString(t *testing.T) {
	balance := Balance{
		Address:          crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
		Amount:           std.NewCoins(std.NewCoin("test", 100)),
		OriginalVesting:  std.NewCoins(std.NewCoin("test", 50)),
		VestingStartTime: 100,
		VestingEndTime:   200,
	}

	str := balance.String()
	assert.Contains(t, str, "100test")
	assert.Contains(t, str, "vesting=50test")
	assert.Contains(t, str, "start=100")
	assert.Contains(t, str, "end=200")
}

func TestBalance_VestingAminoRoundTrip(t *testing.T) {
	expected := Balance{
		Address:          crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
		Amount:           std.MustParseCoins(ugnot.ValueString(100)),
		OriginalVesting:  std.MustParseCoins(ugnot.ValueString(50)),
		VestingStartTime: 100,
		VestingEndTime:   200,
	}

	// Marshal
	value := fmt.Sprintf("[%q]", expected.String())
	var balances []Balance
	err := amino.UnmarshalJSON([]byte(value), &balances)
	require.NoError(t, err)
	require.Len(t, balances, 1)

	balance := balances[0]
	require.Equal(t, expected.Address, balance.Address)
	require.True(t, expected.Amount.IsEqual(balance.Amount))
	require.True(t, expected.OriginalVesting.IsEqual(balance.OriginalVesting))
	require.Equal(t, expected.VestingStartTime, balance.VestingStartTime)
	require.Equal(t, expected.VestingEndTime, balance.VestingEndTime)

	// Marshal back
	balancesJSON, err := amino.MarshalJSON([]Balance{expected})
	require.NoError(t, err)
	require.JSONEq(t, value, string(balancesJSON))
}

func TestBalance_IsVesting(t *testing.T) {
	assert.True(t, Balance{OriginalVesting: std.NewCoins(std.NewCoin("ugnot", 1))}.IsVesting())
	assert.False(t, Balance{}.IsVesting())
}

// -----------------------------------------------------------------------------
// Vesting balance sheet tests

func TestBalances_VestingFromSheet(t *testing.T) {
	t.Parallel()

	dummyKey := getDummyKey(t).PubKey()
	addr := dummyKey.Address().String()

	entries := []string{
		fmt.Sprintf("%s=%s", addr, ugnot.ValueString(100)),
		fmt.Sprintf("%s=%s;vesting=%s;start=100;end=200", addr, ugnot.ValueString(200), ugnot.ValueString(100)),
	}

	balances, err := GetBalancesFromEntries(entries...)
	require.NoError(t, err)

	// The second entry overwrites the first (same address).
	bal := balances[dummyKey.Address()]
	assert.True(t, bal.IsVesting())
	assert.Equal(t, int64(100), bal.VestingStartTime)
	assert.Equal(t, int64(200), bal.VestingEndTime)
	assert.True(t, std.NewCoins(std.NewCoin(ugnot.Denom, 100)).IsEqual(bal.OriginalVesting))
}
