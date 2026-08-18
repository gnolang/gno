package auth

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// TestAddressStoreKeyFormat pins the regular-account key format:
//
//	/a/<20-byte address>
//
// Changing the format requires a coordinated migration of all stored
// accounts and breaks IterateAccounts filtering (which depends on
// AccountStoreKeyLen). Guards against accidental format drift.
func TestAddressStoreKeyFormat(t *testing.T) {
	t.Parallel()

	addr := crypto.AddressFromPreimage([]byte("master"))
	key := AddressStoreKey(addr)

	// Length is exactly AccountStoreKeyLen.
	assert.Equal(t, AccountStoreKeyLen, len(key), "regular account key must be exactly AccountStoreKeyLen bytes")

	// Prefix is "/a/".
	assert.True(t, bytes.HasPrefix(key, []byte(AddressStoreKeyPrefix)), "key must start with /a/")

	// Suffix is the address bytes.
	assert.Equal(t, addr.Bytes(), key[len(AddressStoreKeyPrefix):], "key body must be the 20-byte address")
}

// TestSessionStoreKeyFormat pins the session-account key format:
//
//	/a/<20-byte master>/s/<20-byte session>
//
// The format nests sessions under the master's account prefix so that
// PrefixIterator on "/a/" returns them alongside regular accounts, and
// IterateAccounts uses AccountStoreKeyLen to skip session keys. Any
// change to the format (prefix bytes, infix separator, or byte lengths)
// WILL break IterateAccounts, RemoveAllSessions, IterateSessions, and
// session storage reads/writes across the board. This test is a
// regression guard: update its expected bytes in lockstep with any
// deliberate format change.
func TestSessionStoreKeyFormat(t *testing.T) {
	t.Parallel()

	master := crypto.AddressFromPreimage([]byte("master"))
	session := crypto.AddressFromPreimage([]byte("session"))
	key := SessionStoreKey(master, session)

	// Construct expected key manually: /a/<master>/s/<session>.
	expected := append(append([]byte("/a/"), master.Bytes()...), []byte("/s/")...)
	expected = append(expected, session.Bytes()...)

	assert.Equal(t, expected, key, "session key must match /a/<master>/s/<session> exactly")

	// Length check: strictly greater than AccountStoreKeyLen, which is
	// the property IterateAccounts relies on to skip sessions.
	assert.Greater(t, len(key), AccountStoreKeyLen,
		"session keys must be longer than AccountStoreKeyLen so iteration can filter them out")

	// Explicit layout check via slicing.
	prefix := key[:len(AddressStoreKeyPrefix)]
	masterBytes := key[len(AddressStoreKeyPrefix) : len(AddressStoreKeyPrefix)+crypto.AddressSize]
	infix := key[len(AddressStoreKeyPrefix)+crypto.AddressSize : len(AddressStoreKeyPrefix)+crypto.AddressSize+len(SessionStoreKeyInfix)]
	sessionBytes := key[len(AddressStoreKeyPrefix)+crypto.AddressSize+len(SessionStoreKeyInfix):]

	assert.Equal(t, AddressStoreKeyPrefix, string(prefix))
	assert.Equal(t, master.Bytes(), masterBytes)
	assert.Equal(t, SessionStoreKeyInfix, string(infix))
	assert.Equal(t, session.Bytes(), sessionBytes)
}

// TestSessionPrefixKeyFormat pins the master-session-prefix format used
// for RevokeAllSessions (prefix delete) and IterateSessions iteration.
// Changing the format breaks both operations.
func TestSessionPrefixKeyFormat(t *testing.T) {
	t.Parallel()

	master := crypto.AddressFromPreimage([]byte("master"))
	prefix := SessionPrefixKey(master)

	expected := append(append([]byte("/a/"), master.Bytes()...), []byte("/s/")...)
	assert.Equal(t, expected, prefix)

	// The session prefix must be a proper prefix of every SessionStoreKey
	// for that master — this is what makes prefix delete and iteration
	// correct.
	for _, seed := range []string{"s1", "s2", "sessionWithLongerName"} {
		sessionAddr := crypto.AddressFromPreimage([]byte(seed))
		fullKey := SessionStoreKey(master, sessionAddr)
		require.True(t, bytes.HasPrefix(fullKey, prefix),
			"session key %x must start with session prefix %x", fullKey, prefix)
	}
}

// TestSessionAndRegularAccountKeysDoNotCollide verifies the length-based
// filter used by IterateAccounts. A regular account key for any address
// is exactly AccountStoreKeyLen bytes; a session key for any master and
// session is strictly longer. This test holds for random addresses.
func TestSessionAndRegularAccountKeysDoNotCollide(t *testing.T) {
	t.Parallel()

	for _, seed := range []string{"a", "bc", "longerPreimage123"} {
		addr := crypto.AddressFromPreimage([]byte(seed))
		regKey := AddressStoreKey(addr)
		assert.Equal(t, AccountStoreKeyLen, len(regKey))

		sessionKey := SessionStoreKey(addr, addr)
		assert.NotEqual(t, AccountStoreKeyLen, len(sessionKey),
			"session key must not have the same length as a regular account key")
		assert.Greater(t, len(sessionKey), AccountStoreKeyLen)
	}
}
