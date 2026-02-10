package keys

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto/internal/ledger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateLedgerUnsupportedAlgo(t *testing.T) {
	ledger.Discover = ledger.DiscoverMock
	t.Cleanup(func() { ledger.Discover = ledger.DiscoverDefault })

	kb := NewInMemory()
	_, err := kb.CreateLedger("some_account", Ed25519, "cosmos", 0, 1)
	assert.Error(t, err)
	assert.Equal(t, "unsupported signing algo: only secp256k1 is supported", err.Error())
}

func TestCreateLedger(t *testing.T) {
	ledger.Discover = ledger.DiscoverMock
	t.Cleanup(func() { ledger.Discover = ledger.DiscoverDefault })

	kb := NewInMemory()

	// test_cover and test_unit will result in different answers
	// test_cover does not compile some dependencies so ledger is disabled
	// test_unit may add a ledger mock
	// both cases are acceptable
	_, err := kb.CreateLedger("some_account", Secp256k1, "cosmos", 3, 1)
	require.NoError(t, err)

	// Check that restoring the key gets the same results
	restoredKey, err := kb.GetByName("some_account")
	require.NoError(t, err)
	assert.NotNil(t, restoredKey)
	assert.Equal(t, "some_account", restoredKey.GetName())
	assert.Equal(t, TypeLedger, restoredKey.GetType())

	path, err := restoredKey.GetPath()
	assert.NoError(t, err)
	assert.Equal(t, "44'/118'/3'/0/1", path.String())
}
