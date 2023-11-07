package gnoland

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/jaekwon/testify/assert"
	"github.com/jaekwon/testify/require"
)

func TestBalance_Verify(t *testing.T) {
	validAddress := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	emptyAmount := std.Coins{}
	nonEmptyAmount := std.NewCoins(std.NewCoin("test", 100))

	tests := []struct {
		name      string
		balance   Balance
		expectErr bool
	}{
		{"empty amount", Balance{Address: validAddress, Amount: emptyAmount}, true},
		{"empty address", Balance{Address: bft.Address{}, Amount: nonEmptyAmount}, true},
		{"valid balance", Balance{Address: validAddress, Amount: nonEmptyAmount}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.balance.Verify()
			if tc.expectErr {
				assert.Error(t, err, fmt.Sprintf("TestVerifyBalance: %s", tc.name))
			} else {
				assert.NoError(t, err, fmt.Sprintf("TestVerifyBalance: %s", tc.name))
			}
		})
	}
}

func TestBalance_Parse(t *testing.T) {
	validAddress := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	validBalance := Balance{Address: validAddress, Amount: std.NewCoins(std.NewCoin("test", 100))}

	tests := []struct {
		name      string
		entry     string
		expected  Balance
		expectErr bool
	}{
		{"valid entry", "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=100test", validBalance, false},
		{"invalid address", "invalid=100test", Balance{}, true},
		{"incomplete entry", "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Balance{}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			balance := Balance{}
			err := balance.Parse(tc.entry)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, balance)
			}
		})
	}
}

func TestBalance_AminoUnmarshalJSON(t *testing.T) {
	expected := Balance{
		Address: crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
		Amount:  std.MustParseCoins("100ugnot"),
	}
	value := fmt.Sprintf("[%q]", expected.String())

	var balances []Balance
	err := amino.UnmarshalJSON([]byte(value), &balances)
	require.NoError(t, err)
	require.Len(t, balances, 1, "there should be one balance after unmarshaling")

	balance := balances[0]
	require.Equal(t, expected.Address, balance.Address)
	require.True(t, expected.Amount.IsEqual(balance.Amount))
}

func TestBalance_AminoMarshalJSON(t *testing.T) {
	expected := Balance{
		Address: crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
		Amount:  std.MustParseCoins("100ugnot"),
	}
	expectedJSON := fmt.Sprintf("[%q]", expected.String())

	balancesJSON, err := amino.MarshalJSON([]Balance{expected})
	require.NoError(t, err)
	require.JSONEq(t, expectedJSON, string(balancesJSON))
}
