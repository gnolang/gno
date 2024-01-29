package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecrets_Verify_Single(t *testing.T) {
	t.Parallel()

	t.Run("no individual path set", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"single",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errNoOutputSet.Error())
	})

	t.Run("invalid validator key path", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"single",
			"--validator-key-path",
			"random path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "unable to read validator key")
	})

	t.Run("invalid validator key", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		path := filepath.Join(dirPath, "data.json")

		invalidKey := &privval.FilePVKey{
			PrivKey: nil, // invalid
		}

		require.NoError(t, saveSecretData(invalidKey, path))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"single",
			"--validator-key-path",
			path,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errInvalidPrivateKey)
	})

	t.Run("invalid validator state", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		path := filepath.Join(dirPath, "data.json")

		invalidState := &privval.FilePVLastSignState{
			Height: -1, // invalid
		}

		require.NoError(t, saveSecretData(invalidState, path))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"single",
			"--validator-state-path",
			path,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errInvalidSignStateHeight)
	})

	t.Run("invalid validator state signature", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		keyPath := filepath.Join(dirPath, "key.json")
		statePath := filepath.Join(dirPath, "state.json")

		validKey := generateValidatorPrivateKey()
		validState := generateLastSignValidatorState()

		// Save an invalid signature
		validState.Signature = []byte("totally valid signature")

		require.NoError(t, saveSecretData(validKey, keyPath))
		require.NoError(t, saveSecretData(validState, statePath))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"single",
			"--validator-key-path",
			keyPath,
			"--validator-state-path",
			statePath,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errSignatureValuesMissing)
	})

	t.Run("invalid node key", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		path := filepath.Join(dirPath, "data.json")

		invalidNodeKey := &p2p.NodeKey{
			PrivKey: nil, // invalid
		}

		require.NoError(t, saveSecretData(invalidNodeKey, path))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"single",
			"--node-key-path",
			path,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errInvalidNodeKey)
	})

	t.Run("all secrets set and valid", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		keyPath := filepath.Join(dirPath, "key.json")
		statePath := filepath.Join(dirPath, "state.json")
		nodeKeyPath := filepath.Join(dirPath, "p2p.json")

		validKey := generateValidatorPrivateKey()
		validState := generateLastSignValidatorState()
		validNodeKey := generateNodeKey()

		require.NoError(t, saveSecretData(validKey, keyPath))
		require.NoError(t, saveSecretData(validState, statePath))
		require.NoError(t, saveSecretData(validNodeKey, nodeKeyPath))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"single",
			"--validator-key-path",
			keyPath,
			"--validator-state-path",
			statePath,
			"--node-key-path",
			nodeKeyPath,
		}

		// Run the command
		assert.NoError(t, cmd.ParseAndRun(context.Background(), args))
	})
}
