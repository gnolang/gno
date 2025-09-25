package std

import (
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseSession(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()

	session := NewBaseSession(addr, pubKey, 0)
	require.NotNil(t, session)

	assert.Equal(t, addr, session.GetAddress())
	assert.Equal(t, pubKey, session.GetPubKey())
	assert.Equal(t, uint64(0), session.GetSequence())
	assert.True(t, session.IsActive())
	assert.False(t, session.IsExpired())
	assert.False(t, session.IsFrozen())
}

func TestBaseSessionState(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	session := NewBaseSession(privKey.PubKey().Address(), privKey.PubKey(), 0)

	// Test initial state
	assert.True(t, session.IsActive())
	assert.True(t, session.CanSign([]byte("test")))
	assert.True(t, session.CanExecute("test/path", "method"))
	assert.True(t, session.HasPermission("test.permission"))

	// Test frozen state
	session.Freeze()
	assert.True(t, session.IsFrozen())
	assert.False(t, session.IsActive())
	assert.False(t, session.CanSign([]byte("test")))
	assert.False(t, session.CanExecute("test/path", "method"))
	assert.False(t, session.HasPermission("test.permission"))

	// Test unfreezing
	session.Unfreeze()
	assert.True(t, session.IsActive())
	assert.False(t, session.IsFrozen())

	// Test expiration
	session.Expire()
	assert.True(t, session.IsExpired())
	assert.False(t, session.IsActive())
	assert.False(t, session.CanSign([]byte("test")))
	assert.False(t, session.CanExecute("test/path", "method"))
	assert.False(t, session.HasPermission("test.permission"))
}

func TestBaseSessionSequence(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	session := NewBaseSession(privKey.PubKey().Address(), privKey.PubKey(), 0)

	assert.Equal(t, uint64(0), session.GetSequence())

	session.SetSequence(1)
	assert.Equal(t, uint64(1), session.GetSequence())

	session.SetSequence(100)
	assert.Equal(t, uint64(100), session.GetSequence())
}

func TestBaseSessionLastUsed(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	session := NewBaseSession(privKey.PubKey().Address(), privKey.PubKey(), 0)

	initialLastUsed := session.LastUsedAt
	time.Sleep(time.Millisecond) // Ensure time difference

	session.UpdateLastUsed()
	assert.True(t, session.LastUsedAt.After(initialLastUsed))
}
