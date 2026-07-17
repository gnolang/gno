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

func TestBalance_VestingVerify(t *testing.T) {
	validAddress := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	amount := std.NewCoins(std.NewCoin("test", 100))
	vestingAmt := std.NewCoins(std.NewCoin("test", 50))

	tests := []struct {
		name      string
		balance   Balance
		expectErr bool
	}{
		{
			"valid vesting",
			Balance{
				Address: validAddress, Amount: amount,
				Vesting: &std.VestingSchedule{OriginalVesting: vestingAmt, StartTime: 100, EndTime: 200},
			},
			false,
		},
		{
			"vesting start >= end",
			Balance{
				Address: validAddress, Amount: amount,
				Vesting: &std.VestingSchedule{OriginalVesting: vestingAmt, StartTime: 200, EndTime: 100},
			},
			true,
		},
		{
			"vesting exceeds balance",
			Balance{
				Address: validAddress, Amount: vestingAmt,
				Vesting: &std.VestingSchedule{OriginalVesting: amount, StartTime: 100, EndTime: 200},
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

func TestBalance_VestingParse(t *testing.T) {
	validAddress := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	tests := []struct {
		name      string
		entry     string
		expected  Balance
		expectErr bool
	}{
		{
			"no vesting",
			"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=100test",
			Balance{Address: validAddress, Amount: std.NewCoins(std.NewCoin("test", 100))},
			false,
		},
		{
			"with vesting (continuous)",
			"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=100test;vesting=50test,100,200",
			Balance{
				Address: validAddress,
				Amount:  std.NewCoins(std.NewCoin("test", 100)),
				Vesting: &std.VestingSchedule{
					OriginalVesting: std.NewCoins(std.NewCoin("test", 50)),
					StartTime:       100,
					EndTime:         200,
				},
			},
			false,
		},
		{
			"with vesting (delayed)",
			"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=100test;vesting=50test,0,200;type=delayed",
			Balance{
				Address: validAddress,
				Amount:  std.NewCoins(std.NewCoin("test", 100)),
				Vesting: &std.VestingSchedule{
					OriginalVesting: std.NewCoins(std.NewCoin("test", 50)),
					EndTime:         200,
					Type:            std.VestingDelayed,
				},
			},
			false,
		},
		{
			"invalid vesting amount",
			"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=100test;vesting=invalid,100,200",
			Balance{},
			true,
		},
		{
			"invalid vesting start time",
			"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=100test;vesting=50test,abc,200",
			Balance{},
			true,
		},
		{
			"malformed vesting (not enough fields)",
			"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=100test;vesting=50test,100",
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
				if tc.expected.Vesting != nil {
					require.NotNil(t, balance.Vesting)
					assert.True(t, tc.expected.Vesting.OriginalVesting.IsEqual(balance.Vesting.OriginalVesting))
					assert.Equal(t, tc.expected.Vesting.StartTime, balance.Vesting.StartTime)
					assert.Equal(t, tc.expected.Vesting.EndTime, balance.Vesting.EndTime)
				} else {
					assert.Nil(t, balance.Vesting)
				}
			}
		})
	}
}

func TestBalance_VestingString(t *testing.T) {
	balance := Balance{
		Address: crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
		Amount:  std.NewCoins(std.NewCoin("test", 100)),
		Vesting: &std.VestingSchedule{
			OriginalVesting: std.NewCoins(std.NewCoin("test", 50)),
			StartTime:       100,
			EndTime:         200,
		},
	}

	str := balance.String()
	assert.Contains(t, str, "100test")
	assert.Contains(t, str, "vesting=50test,100,200")

	// Round-trip
	var parsed Balance
	err := parsed.Parse(str)
	require.NoError(t, err)
	assert.Equal(t, balance.Address, parsed.Address)
	assert.True(t, balance.Amount.IsEqual(parsed.Amount))
	require.NotNil(t, parsed.Vesting)
	assert.True(t, balance.Vesting.OriginalVesting.IsEqual(parsed.Vesting.OriginalVesting))
	assert.Equal(t, balance.Vesting.StartTime, parsed.Vesting.StartTime)
	assert.Equal(t, balance.Vesting.EndTime, parsed.Vesting.EndTime)
}

func TestBalance_VestingAminoRoundTrip(t *testing.T) {
	expected := Balance{
		Address: crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
		Amount:  std.MustParseCoins(ugnot.ValueString(100)),
		Vesting: &std.VestingSchedule{
			OriginalVesting: std.MustParseCoins(ugnot.ValueString(50)),
			StartTime:       100,
			EndTime:         200,
		},
	}

	value := fmt.Sprintf("[%q]", expected.String())
	var balances []Balance
	err := amino.UnmarshalJSON([]byte(value), &balances)
	require.NoError(t, err)
	require.Len(t, balances, 1)

	balance := balances[0]
	require.Equal(t, expected.Address, balance.Address)
	require.True(t, expected.Amount.IsEqual(balance.Amount))
	require.NotNil(t, balance.Vesting)
	require.True(t, expected.Vesting.OriginalVesting.IsEqual(balance.Vesting.OriginalVesting))
	require.Equal(t, expected.Vesting.StartTime, balance.Vesting.StartTime)
	require.Equal(t, expected.Vesting.EndTime, balance.Vesting.EndTime)

	balancesJSON, err := amino.MarshalJSON([]Balance{expected})
	require.NoError(t, err)
	require.JSONEq(t, value, string(balancesJSON))
}

func TestBalance_IsVesting(t *testing.T) {
	assert.True(t, Balance{
		Vesting: &std.VestingSchedule{OriginalVesting: std.NewCoins(std.NewCoin("ugnot", 1))},
	}.IsVesting())

	assert.False(t, Balance{}.IsVesting())
	assert.False(t, Balance{
		Vesting: &std.VestingSchedule{},
	}.IsVesting())
}

func TestBalances_VestingFromEntries(t *testing.T) {
	t.Parallel()

	dummyKey := getDummyKey(t).PubKey()
	addr := dummyKey.Address().String()

	entries := []string{
		fmt.Sprintf("%s=%s", addr, ugnot.ValueString(200)),
		fmt.Sprintf("%s=%s;vesting=%s,100,200", addr, ugnot.ValueString(300), ugnot.ValueString(100)),
	}

	balances, err := GetBalancesFromEntries(entries...)
	require.NoError(t, err)

	bal := balances[dummyKey.Address()]
	assert.True(t, bal.IsVesting())
	assert.Equal(t, int64(100), bal.Vesting.StartTime)
	assert.Equal(t, int64(200), bal.Vesting.EndTime)
	assert.True(t, std.NewCoins(std.NewCoin(ugnot.Denom, 100)).IsEqual(bal.Vesting.OriginalVesting))
}
