package types

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateKeys generates random node p2p keys
func generateKeys(t *testing.T, count int) []*NodeKey {
	t.Helper()

	keys := make([]*NodeKey, count)

	for i := range count {
		keys[i] = GenerateNodeKey()
	}

	return keys
}

func TestNodeKey_Generate(t *testing.T) {
	t.Parallel()

	keys := generateKeys(t, 10)

	for _, key := range keys {
		require.NotNil(t, key)
		assert.NotNil(t, key.PrivKey)

		// Make sure all keys are unique
		for _, keyInner := range keys {
			if key.ID() == keyInner.ID() {
				continue
			}

			assert.False(t, key.PrivKey.Equals(keyInner.PrivKey))
		}
	}
}

func TestNodeKey_Load(t *testing.T) {
	t.Parallel()

	t.Run("non-existing key", func(t *testing.T) {
		t.Parallel()

		key, err := LoadNodeKey("definitely valid path")

		require.Nil(t, key)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("invalid key format", func(t *testing.T) {
		t.Parallel()

		// Generate a random path
		path := fmt.Sprintf("%s/key.json", t.TempDir())

		type random struct {
			Field string
		}

		data, err := json.Marshal(&random{
			Field: "random data",
		})
		require.NoError(t, err)

		// Save the invalid data format
		require.NoError(t, os.WriteFile(path, data, 0o644))

		// Load the key, that's invalid
		_, err = LoadNodeKey(path)
		require.ErrorIs(t, err, errInvalidNodeKey)
	})

	t.Run("valid key loaded", func(t *testing.T) {
		t.Parallel()

		path := fmt.Sprintf("%s/key.json", t.TempDir())

		// Save the key
		key, err := GeneratePersistedNodeKey(path)
		require.NoError(t, err)

		// Load the key, that's valid
		loadedKey, err := LoadNodeKey(path)
		require.NoError(t, err)

		assert.True(t, key.PrivKey.Equals(loadedKey.PrivKey))
		assert.Equal(t, key.ID(), loadedKey.ID())
	})
}

func TestNodeKey_ID(t *testing.T) {
	t.Parallel()

	keys := generateKeys(t, 10)

	for _, key := range keys {
		// Make sure the ID is valid
		id := key.ID()
		require.NotNil(t, id)

		assert.NoError(t, id.Validate())
	}
}

func TestNodeKey_LoadOrGenNodeKey(t *testing.T) {
	t.Parallel()

	t.Run("existing key loaded", func(t *testing.T) {
		t.Parallel()

		path := fmt.Sprintf("%s/key.json", t.TempDir())

		// Save the key
		key, err := GeneratePersistedNodeKey(path)
		require.NoError(t, err)

		// Load the saved key
		loadedKey, err := LoadNodeKey(path)
		require.NoError(t, err)

		// Make sure the key was not generated
		assert.True(t, key.PrivKey.Equals(loadedKey.PrivKey))
	})

	t.Run("fresh key generated", func(t *testing.T) {
		t.Parallel()

		path := fmt.Sprintf("%s/key.json", t.TempDir())

		// Make sure there is no key at the path
		_, err := os.Stat(path)
		require.ErrorIs(t, err, os.ErrNotExist)

		// Generate the fresh key
		key, err := GeneratePersistedNodeKey(path)
		require.NoError(t, err)

		// Load the saved key
		loadedKey, err := LoadNodeKey(path)
		require.NoError(t, err)

		// Make sure the keys are the same
		assert.True(t, key.PrivKey.Equals(loadedKey.PrivKey))
	})
}
