package std

import (
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/require"
)

func TestSessionManager(t *testing.T) {
	config := SessionConfig{
		MaxSessionsPerAccount: 2,
		ExpirationDuration:    1 * time.Hour,
		CleanupInterval:       1 * time.Minute,
	}

	sm := NewSessionManager(config)
	require.NotNil(t, sm)

	// test address
	privKey1 := secp256k1.GenPrivKey()
	pubKey1 := privKey1.PubKey()
	addr1 := pubKey1.Address()

	privKey2 := secp256k1.GenPrivKey()
	pubKey2 := privKey2.PubKey()

	t.Run("CreateSession", func(t *testing.T) {
		// first session creation
		session1, err := sm.CreateSession(addr1, pubKey1)
		require.NoError(t, err)
		require.NotNil(t, session1)
		require.Equal(t, addr1, session1.Address)
		require.True(t, pubKey1.Equals(session1.PubKey))
		require.False(t, session1.ExpirationTime.IsZero())

		// second session creation
		session2, err := sm.CreateSession(addr1, pubKey2)
		require.NoError(t, err)
		require.NotNil(t, session2)

		// over max session count test
		privKey3 := secp256k1.GenPrivKey()
		pubKey3 := privKey3.PubKey()
		session3, err := sm.CreateSession(addr1, pubKey3)
		require.Error(t, err)
		require.Equal(t, errTooManySessions, err)
		require.Nil(t, session3)
	})

	t.Run("GetSession", func(t *testing.T) {
		// existing session get
		session, err := sm.GetSession(pubKey1)
		require.NoError(t, err)
		require.NotNil(t, session)
		require.True(t, pubKey1.Equals(session.PubKey))

		// non-existent session get
		privKeyNonExistent := secp256k1.GenPrivKey()
		pubKeyNonExistent := privKeyNonExistent.PubKey()
		session, err = sm.GetSession(pubKeyNonExistent)
		require.Error(t, err)
		require.Equal(t, errSessionNotFound, err)
		require.Nil(t, session)
	})

	t.Run("GetSessionsByAddress", func(t *testing.T) {
		sessions := sm.GetSessionsByAddress(addr1)
		require.Len(t, sessions, 2)
	})

	t.Run("RemoveSession", func(t *testing.T) {
		err := sm.RemoveSession(pubKey1)
		require.NoError(t, err)

		// search removed session
		session, err := sm.GetSession(pubKey1)
		require.Error(t, err)
		require.Equal(t, errSessionNotFound, err)
		require.Nil(t, session)

		// remove non-existent session
		err = sm.RemoveSession(pubKey1)
		require.Error(t, err)
		require.Equal(t, errSessionNotFound, err)
	})

	t.Run("UpdateSession", func(t *testing.T) {
		// update session
		err := sm.UpdateSession(pubKey2, func(s *BaseSession) error {
			s.State = SessionStateFrozen
			return nil
		})
		require.NoError(t, err)

		// check updated session
		session, err := sm.GetSession(pubKey2)
		require.NoError(t, err)
		require.Equal(t, SessionStateFrozen, session.State)
	})
}

func TestSessionExpiration(t *testing.T) {
	config := SessionConfig{
		MaxSessionsPerAccount: 2,
		ExpirationDuration:    10 * time.Millisecond,
		CleanupInterval:       20 * time.Millisecond,
	}

	sm := NewSessionManager(config)

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()

	// session creation
	session, err := sm.CreateSession(addr, pubKey)
	require.NoError(t, err)
	require.NotNil(t, session)

	// before expiration session get
	session, err = sm.GetSession(pubKey)
	require.NoError(t, err)
	require.NotNil(t, session)

	// wait for expiration
	time.Sleep(15 * time.Millisecond)

	// expired session get
	session, err = sm.GetSession(pubKey)
	require.Error(t, err)
	require.Equal(t, errSessionExpired, err)
	require.Nil(t, session)

	// wait for cleanup
	time.Sleep(25 * time.Millisecond)

	// check session is cleaned up
	sessions := sm.GetSessionsByAddress(addr)
	require.Empty(t, sessions)
}
