package crypto_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

var invalidStrs = []string{
	crypto.Bech32AddrPrefix + "AB0C",
	crypto.Bech32AddrPrefix + "1234",
	crypto.Bech32AddrPrefix + "5678",
	crypto.Bech32AddrPrefix + "BBAB",
	crypto.Bech32AddrPrefix + "FF04",
	crypto.Bech32AddrPrefix + "6789",
}

func TestEmptyAddresses(t *testing.T) {
	require.Equal(t, (crypto.Address{}).String(), "g1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqluuxe")

	addr := crypto.AddressFromBytes(make([]byte, 20))
	require.True(t, addr.IsZero())

	addr, err := crypto.AddressFromBech32("")
	require.True(t, addr.IsZero())
	require.NotNil(t, err)

	addr, err = crypto.AddressFromBech32("g1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqluuxe")
	require.True(t, addr.IsZero())
	require.Nil(t, err)
}

func testMarshal(t *testing.T, addr crypto.Address, marshal func(orig interface{}) ([]byte, error), unmarshal func(bz []byte, ptr interface{}) error) {
	t.Helper()

	bz, err := marshal(addr)
	require.Nil(t, err)
	res := crypto.Address{}
	err = unmarshal(bz, &res)
	require.Nil(t, err)
	require.Equal(t, addr, res)
}

func TestRandBech32AddrConsistency(t *testing.T) {
	var pub ed25519.PubKeyEd25519

	for i := 0; i < 1000; i++ {
		rand.Read(pub[:])

		addr := crypto.AddressFromBytes(pub.Address().Bytes())
		testMarshal(t, addr, amino.Marshal, amino.Unmarshal)
		testMarshal(t, addr, amino.MarshalJSON, amino.UnmarshalJSON)

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
