package balances

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/jaekwon/testify/require"
	"github.com/stretchr/testify/assert"
)

func TestBalances_GetBalancesFromEntries(t *testing.T) {
	t.Parallel()

	t.Run("valid balances", func(t *testing.T) {
		t.Parallel()

		// Generate dummy keys
		dummyKeys := getDummyKeys(t, 2)
		amount := std.NewCoins(std.NewCoin("ugnot", 10))

		entries := make([]string, len(dummyKeys))

		for index, key := range dummyKeys {
			entries[index] = fmt.Sprintf(
				"%s=%dugnot",
				key.Address().String(),
				amount.AmountOf("ugnot"),
			)
		}

		balanceMap, err := GetBalancesFromEntries(entries...)
		require.NoError(t, err)

		// Validate the balance map
		assert.Len(t, balanceMap, len(dummyKeys))
		for _, key := range dummyKeys {
			assert.Equal(t, amount, balanceMap[key.Address()].Amount)
		}
	})

	t.Run("malformed balance, invalid format", func(t *testing.T) {
		t.Parallel()

		entries := []string{
			"malformed balance",
		}

		balanceMap, err := GetBalancesFromEntries(entries...)

		assert.Nil(t, balanceMap)
		assert.ErrorContains(t, err, "malformed entry")
	})

	t.Run("malformed balance, invalid address", func(t *testing.T) {
		t.Parallel()

		balances := []string{
			"dummyaddress=10ugnot",
		}

		balanceMap, err := GetBalancesFromEntries(balances...)

		assert.Nil(t, balanceMap)
		assert.ErrorContains(t, err, "invalid address")
	})

	t.Run("malformed balance, invalid amount", func(t *testing.T) {
		t.Parallel()

		dummyKey := getDummyKey(t)

		balances := []string{
			fmt.Sprintf(
				"%s=%sugnot",
				dummyKey.Address().String(),
				strconv.FormatUint(math.MaxUint64, 10),
			),
		}

		balanceMap, err := GetBalancesFromEntries(balances...)

		assert.Nil(t, balanceMap)
		assert.ErrorContains(t, err, "invalid amount")
	})
}

func TestBalances_GetBalancesFromSheet(t *testing.T) {
	t.Parallel()

	t.Run("valid balances", func(t *testing.T) {
		t.Parallel()

		// Generate dummy keys
		dummyKeys := getDummyKeys(t, 2)
		amount := std.NewCoins(std.NewCoin("ugnot", 10))

		balances := make([]string, len(dummyKeys))

		for index, key := range dummyKeys {
			balances[index] = fmt.Sprintf(
				"%s=%dugnot",
				key.Address().String(),
				amount.AmountOf("ugnot"),
			)
		}

		reader := strings.NewReader(strings.Join(balances, "\n"))
		balanceMap, err := GetBalancesFromSheet(reader)
		require.NoError(t, err)

		// Validate the balance map
		assert.Len(t, balanceMap, len(dummyKeys))
		for _, key := range dummyKeys {
			assert.Equal(t, amount, balanceMap[key.Address()].Amount)
		}
	})

	t.Run("malformed balance, invalid amount", func(t *testing.T) {
		t.Parallel()

		dummyKey := getDummyKey(t)

		balances := []string{
			fmt.Sprintf(
				"%s=%sugnot",
				dummyKey.Address().String(),
				strconv.FormatUint(math.MaxUint64, 10),
			),
		}

		reader := strings.NewReader(strings.Join(balances, "\n"))

		balanceMap, err := GetBalancesFromSheet(reader)

		assert.Nil(t, balanceMap)
		assert.ErrorContains(t, err, "invalid amount")
	})
}

// XXX: this function should probably be exposed somewhere as it's duplicate of
// cmd/genesis/...

// getDummyKey generates a random public key,
// and returns the key info
func getDummyKey(t *testing.T) crypto.PubKey {
	t.Helper()

	mnemonic, err := client.GenerateMnemonic(256)
	require.NoError(t, err)

	seed := bip39.NewSeed(mnemonic, "")

	return generateKeyFromSeed(seed, 0).PubKey()
}

// getDummyKeys generates random keys for testing
func getDummyKeys(t *testing.T, count int) []crypto.PubKey {
	t.Helper()

	dummyKeys := make([]crypto.PubKey, count)

	for i := 0; i < count; i++ {
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
