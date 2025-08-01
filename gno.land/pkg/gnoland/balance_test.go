package gnoland

import (
	crand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/amino"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		Amount:  std.MustParseCoins(ugnot.ValueString(100)),
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
		Amount:  std.MustParseCoins(ugnot.ValueString(100)),
	}
	expectedJSON := fmt.Sprintf("[%q]", expected.String())

	balancesJSON, err := amino.MarshalJSON([]Balance{expected})
	require.NoError(t, err)
	require.JSONEq(t, expectedJSON, string(balancesJSON))
}

func TestBalances_GetBalancesFromEntries(t *testing.T) {
	t.Parallel()

	t.Run("valid balances", func(t *testing.T) {
		t.Parallel()

		// Generate dummy keys
		dummyKeys := getDummyKeys(t, 2)
		amount := std.NewCoins(std.NewCoin(ugnot.Denom, 10))

		entries := make([]string, len(dummyKeys))

		for index, key := range dummyKeys {
			entries[index] = fmt.Sprintf(
				"%s=%s",
				key.PubKey().Address().String(),
				ugnot.ValueString(amount.AmountOf(ugnot.Denom)),
			)
		}

		balanceMap, err := GetBalancesFromEntries(entries...)
		require.NoError(t, err)

		// Validate the balance map
		assert.Len(t, balanceMap, len(dummyKeys))
		for _, key := range dummyKeys {
			assert.Equal(t, amount, balanceMap[key.PubKey().Address()].Amount)
		}
	})

	t.Run("malformed balance, invalid format", func(t *testing.T) {
		t.Parallel()

		entries := []string{
			"malformed balance",
		}

		balanceMap, err := GetBalancesFromEntries(entries...)
		assert.Len(t, balanceMap, 0)
		assert.Contains(t, err.Error(), "malformed entry")
	})

	t.Run("malformed balance, invalid address", func(t *testing.T) {
		t.Parallel()

		balances := []string{
			"dummyaddress=" + ugnot.ValueString(10),
		}

		balanceMap, err := GetBalancesFromEntries(balances...)
		assert.Len(t, balanceMap, 0)
		assert.ErrorContains(t, err, "invalid address")
	})

	t.Run("malformed balance, invalid amount", func(t *testing.T) {
		t.Parallel()

		dummyKey := getDummyKey(t).PubKey()

		balances := []string{
			fmt.Sprintf(
				"%s=%s%s",
				dummyKey.Address().String(),
				strconv.FormatUint(math.MaxUint64, 10),
				ugnot.Denom,
			),
		}

		balanceMap, err := GetBalancesFromEntries(balances...)
		assert.Len(t, balanceMap, 0)
		assert.ErrorContains(t, err, "invalid amount")
	})
}

func TestBalances_GetBalancesFromSheet(t *testing.T) {
	t.Parallel()

	t.Run("valid balances", func(t *testing.T) {
		t.Parallel()

		// Generate dummy keys
		dummyKeys := getDummyKeys(t, 2)
		amount := std.NewCoins(std.NewCoin(ugnot.Denom, 10))

		balances := make([]string, len(dummyKeys))

		for index, key := range dummyKeys {
			balances[index] = fmt.Sprintf(
				"%s=%s",
				key.PubKey().Address().String(),
				ugnot.ValueString(amount.AmountOf(ugnot.Denom)),
			)
		}

		reader := strings.NewReader(strings.Join(balances, "\n"))
		balanceMap, err := GetBalancesFromSheet(reader)
		require.NoError(t, err)

		// Validate the balance map
		assert.Len(t, balanceMap, len(dummyKeys))
		for _, key := range dummyKeys {
			assert.Equal(t, amount, balanceMap[key.PubKey().Address()].Amount)
		}
	})

	t.Run("malformed balance, invalid amount", func(t *testing.T) {
		t.Parallel()

		dummyKey := getDummyKey(t).PubKey()

		balances := []string{
			fmt.Sprintf(
				"%s=%s%s",
				dummyKey.Address().String(),
				strconv.FormatUint(math.MaxUint64, 10),
				ugnot.Denom,
			),
		}

		reader := strings.NewReader(strings.Join(balances, "\n"))

		balanceMap, err := GetBalancesFromSheet(reader)

		assert.Len(t, balanceMap, 0)
		assert.Contains(t, err.Error(), "invalid amount")
	})
}

func TestBalances_List(t *testing.T) {
	t.Parallel()

	generateBalances := func(count int) Balances {
		balances := NewBalances()

		for i := 0; i < count; i++ {
			// Generate a random address
			var addr bft.Address
			_, err := crand.Read(addr[:])
			require.NoError(t, err)

			// Generate a random amount
			randAmount, err := crand.Int(crand.Reader, big.NewInt(100))
			require.NoError(t, err)

			amount := std.NewCoins(std.NewCoin(ugnot.Denom, randAmount.Int64()+1))
			balances.Set(addr, amount)
		}

		return balances
	}

	t.Run("sorted balance list", func(t *testing.T) {
		t.Parallel()

		const nAccounts = 100

		balances := generateBalances(nAccounts)

		got := balances.List()

		for i := 1; i < len(got); i++ {
			var (
				prev = got[i-1].Address
				curr = got[i].Address
			)

			assert.LessOrEqual(
				t,
				prev.Compare(curr),
				0,
			)
		}
	})

	t.Run("deterministic balance sort", func(t *testing.T) {
		t.Parallel()

		const (
			nAccounts   = 100
			nIterations = 300
		)

		var (
			balances       = generateBalances(nAccounts)
			sortedBalances = make([][]Balance, 0, nIterations)
		)

		for range nIterations {
			sortedBalances = append(sortedBalances, balances.List())
		}

		ref := sortedBalances[0]
		for _, sb := range sortedBalances[1:] {
			require.True(
				t,
				slices.EqualFunc(ref, sb, func(b1 Balance, b2 Balance) bool {
					return b1.Amount.IsEqual(b2.Amount) && b1.Address.Compare(b2.Address) == 0
				}),
			)
		}
	})
}

// XXX: this function should probably be exposed somewhere as it's duplicate of
// cmd/genesis/...

// getDummyKey generates a random private key
func getDummyKey(t *testing.T) crypto.PrivKey {
	t.Helper()

	mnemonic, err := client.GenerateMnemonic(256)
	require.NoError(t, err)

	seed := bip39.NewSeed(mnemonic, "")

	return generateKeyFromSeed(seed, 0)
}

// getDummyKeys generates random keys for testing
func getDummyKeys(t *testing.T, count int) []crypto.PrivKey {
	t.Helper()

	dummyKeys := make([]crypto.PrivKey, count)

	for i := range count {
		dummyKeys[i] = getDummyKey(t)
	}

	return dummyKeys
}

// generateKeyFromSeed generates a private key from
// the provided seed and index
func generateKeyFromSeed(seed []byte, index uint32) crypto.PrivKey {
	pathParams := hd.NewFundraiserParams(0, crypto.CoinType, index)

	masterPriv, ch := hd.ComputeMastersFromSeed(seed)

	//nolint:errcheck // This derivation can never error out, since the path params
	// are always going to be valid
	derivedPriv, _ := hd.DerivePrivateKeyForPath(masterPriv, ch, pathParams.String())

	return secp256k1.PrivKeySecp256k1(derivedPriv)
}
