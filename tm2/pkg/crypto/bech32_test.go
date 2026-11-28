package crypto_test

import (
	"crypto/sha256"
	"encoding/json"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bech32"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

var invalidStrs = []string{
	crypto.Bech32AddrPrefix() + "AB0C",
	crypto.Bech32AddrPrefix() + "1234",
	crypto.Bech32AddrPrefix() + "5678",
	crypto.Bech32AddrPrefix() + "BBAB",
	crypto.Bech32AddrPrefix() + "FF04",
	crypto.Bech32AddrPrefix() + "6789",
}

func TestEmptyAddresses(t *testing.T) {
	t.Parallel()

	require.Equal(t, (crypto.Address{}).String(), "g1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqluuxe")

	addr := crypto.MustAddressFromBytes(make([]byte, 20))
	require.True(t, addr.IsZero())

	addr, err := crypto.AddressFromBech32("")
	require.True(t, addr.IsZero())
	require.NotNil(t, err)

	addr, err = crypto.AddressFromBech32("g1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqluuxe")
	require.True(t, addr.IsZero())
	require.Nil(t, err)
}

func testMarshal(t *testing.T, addr crypto.Address, marshal func(orig any) ([]byte, error), unmarshal func(bz []byte, ptr any) error) {
	t.Helper()

	bz, err := marshal(addr)
	require.Nil(t, err)
	res := crypto.Address{}
	err = unmarshal(bz, &res)
	require.Nil(t, err)
	require.Equal(t, addr, res)
}

func TestRandBech32AddrConsistency(t *testing.T) {
	t.Parallel()

	var pub ed25519.PubKeyEd25519
	cc8 := rand.NewChaCha8(sha256.Sum256([]byte("abc123")))

	for range 1000 {
		cc8.Read(pub[:])

		addr := crypto.MustAddressFromBytes(pub.Address().Bytes())
		testMarshal(t, addr, amino.Marshal, amino.Unmarshal)
		testMarshal(t, addr, amino.MarshalJSON, amino.UnmarshalJSON)
		testMarshal(t, addr, json.Marshal, json.Unmarshal)

		str := addr.String()
		res, err := crypto.AddressFromBech32(str)
		require.Nil(t, err)
		require.Equal(t, addr, res)
	}

	for _, str := range invalidStrs {
		_, err := crypto.AddressFromBech32(str)
		require.NotNil(t, err)
	}
}

// A prefix error names a prefix of at most 256 bytes, and gives a longer one,
// which a peer can make as long as its message, by its length.
func TestPrefixErrorNamesOnlyShortPrefix(t *testing.T) {
	t.Parallel()

	encode := func(hrp string) string {
		s, err := bech32.Encode(hrp, make([]byte, crypto.AddressSize))
		require.NoError(t, err)
		return s
	}
	short, long := encode("x"), encode(strings.Repeat("x", 257))
	want := crypto.Bech32AddrPrefix()

	_, err := crypto.AddressFromBech32(short)
	require.EqualError(t, err, "invalid Bech32 prefix; expected "+want+", got x")
	_, err = crypto.AddressFromBech32(long)
	require.EqualError(t, err, "invalid Bech32 prefix; expected "+want+", got a 257-byte prefix")

	var addr crypto.Address
	require.EqualError(t, addr.DecodeString(short), `unexpected bech32 prefix for address. expected "`+want+`", got "x"`)
	require.EqualError(t, addr.DecodeString(long), `unexpected bech32 prefix for address. expected "`+want+`", got a 257-byte prefix`)
}
