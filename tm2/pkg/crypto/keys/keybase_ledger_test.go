package keys

import (
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/gnolang/gno/tm2/pkg/crypto/internal/ledger"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
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

// TestSignLedgerMismatchedDevice asserts that signing fails when the connected
// Ledger device exposes a different pubkey than the one stored in the keybase.
func TestSignLedgerMismatchedDevice(t *testing.T) {
	var device ledger.SECP256K1
	ledger.Discover = func() (ledger.SECP256K1, error) { return device, nil }
	t.Cleanup(func() { ledger.Discover = ledger.DiscoverDefault })

	// Device A is plugged in when the key reference is created.
	device = newSwappableLedgerMock(t)

	kb := NewInMemory()
	_, err := kb.CreateLedger("acct", Secp256k1, "cosmos", 0, 0)
	require.NoError(t, err)

	// Swap in device B with a different pubkey at the same BIP44 path.
	device = newSwappableLedgerMock(t)

	_, _, err = kb.Sign("acct", "", []byte("hello"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not match the stored public key")
}

type swappableLedgerMock struct {
	ledger.MockLedger
	pubKey []byte
	addr   string
}

func newSwappableLedgerMock(t *testing.T) swappableLedgerMock {
	t.Helper()
	priv := secp256k1.GenPrivKey()
	_, btcPub := btcec.PrivKeyFromBytes(priv[:])
	return swappableLedgerMock{
		pubKey: btcPub.SerializeCompressed(),
		addr:   priv.PubKey().Address().String(),
	}
}

func (m swappableLedgerMock) GetPublicKeySECP256K1(_ []uint32) ([]byte, error) {
	return m.pubKey, nil
}

func (m swappableLedgerMock) GetAddressPubKeySECP256K1(_ []uint32, _ string) ([]byte, string, error) {
	return m.pubKey, m.addr, nil
}
