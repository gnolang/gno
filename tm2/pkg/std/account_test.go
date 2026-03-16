package std

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseSessionAccount_AminoRoundTrip(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	masterAddr := crypto.AddressFromPreimage([]byte("master"))

	original := &BaseSessionAccount{
		BaseAccount: BaseAccount{
			Address:       addr,
			PubKey:        pubKey,
			AccountNumber: 42,
			Sequence:      7,
			// Coins intentionally omitted — session accounts do not hold coins.
		},
		MasterAddress: masterAddr,
		ExpiresAt:     1700000000,
		SpendLimit:    Coins{NewCoin("ugnot", 5000)},
		SpendPeriod:   86400,
		SpendUsed:     Coins{NewCoin("ugnot", 200)},
		SpendReset:    1699990000,
	}

	// Marshal
	bz, err := amino.MarshalAny(original)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	// Unmarshal
	var got interface{}
	err = amino.UnmarshalAny(bz, &got)
	require.NoError(t, err)

	result, ok := got.(*BaseSessionAccount)
	require.True(t, ok, "expected *BaseSessionAccount, got %T", got)

	// Verify all fields
	assert.Equal(t, original.Address, result.Address)
	assert.Nil(t, result.GetCoins(), "session accounts should not hold coins")
	assert.True(t, original.PubKey.Equals(result.PubKey))
	assert.Equal(t, original.AccountNumber, result.AccountNumber)
	assert.Equal(t, original.Sequence, result.Sequence)
	assert.Equal(t, original.MasterAddress, result.MasterAddress)
	assert.Equal(t, original.ExpiresAt, result.ExpiresAt)
	assert.True(t, original.SpendLimit.IsEqual(result.SpendLimit))
	assert.Equal(t, original.SpendPeriod, result.SpendPeriod)
	assert.True(t, original.SpendUsed.IsEqual(result.SpendUsed))
	assert.Equal(t, original.SpendReset, result.SpendReset)
}
