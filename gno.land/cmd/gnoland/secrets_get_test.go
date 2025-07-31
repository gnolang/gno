package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	fstate "github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecrets_Get_All(t *testing.T) {
	t.Parallel()

	t.Run("invalid data directory", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			"",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errInvalidDataDir.Error())
	})

	t.Run("all secrets shown", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		tempDir := t.TempDir()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())

		// Run the init command
		initArgs := []string{
			"secrets",
			"init",
			"--data-dir",
			tempDir,
		}

		// Run the init command
		require.NoError(t, cmd.ParseAndRun(context.Background(), initArgs))

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		cmd = newRootCmd(io)

		// Get the node key
		nodeKeyPath := filepath.Join(tempDir, defaultNodeKeyName)
		nodeKey, err := types.LoadNodeKey(nodeKeyPath)
		require.NoError(t, err)

		// Get the validator private key
		validatorKeyPath := filepath.Join(tempDir, defaultValidatorKeyName)
		validatorKey, err := signer.LoadFileKey(validatorKeyPath)
		require.NoError(t, err)

		// Get the validator state
		validatorStatePath := filepath.Join(tempDir, defaultValidatorStateName)
		state, err := fstate.LoadFileState(validatorStatePath)
		require.NoError(t, err)

		// Run the show command
		showArgs := []string{
			"secrets",
			"get",
			"--data-dir",
			tempDir,
		}

		require.NoError(t, cmd.ParseAndRun(context.Background(), showArgs))

		output := mockOutput.String()

		// Make sure the node p2p key is displayed
		assert.Contains(
			t,
			output,
			nodeKey.ID().String(),
		)

		// Make sure the node p2p address is displayed
		assert.Contains(
			t,
			output,
			constructP2PAddress(nodeKey.ID(), config.DefaultConfig().P2P.ListenAddress),
		)

		// Make sure the private key info is displayed
		assert.Contains(
			t,
			output,
			validatorKey.Address.String(),
		)

		assert.Contains(
			t,
			output,
			validatorKey.PubKey.String(),
		)

		// Make sure the private key info is displayed
		assert.Contains(
			t,
			output,
			validatorKey.Address.String(),
		)

		assert.Contains(
			t,
			output,
			validatorKey.PubKey.String(),
		)

		// Make sure the state info is displayed
		assert.Contains(
			t,
			output,
			fmt.Sprintf("%d", state.Step),
		)

		assert.Contains(
			t,
			output,
			fmt.Sprintf("%d", state.Height),
		)

		assert.Contains(
			t,
			output,
			strconv.Itoa(state.Round),
		)
	})
}

func TestSecrets_Get_ValidatorKeyInfo(t *testing.T) {
	t.Parallel()

	t.Run("validator key info", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		keyPath := filepath.Join(dirPath, defaultValidatorKeyName)

		validKey, err := signer.GeneratePersistedFileKey(keyPath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			dirPath,
			validatorPrivateKeyKey,
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		var vk validatorKeyInfo

		require.NoError(t, json.Unmarshal(mockOutput.Bytes(), &vk))

		// Make sure the private key info is displayed
		assert.Equal(
			t,
			validKey.Address.String(),
			vk.Address,
		)

		assert.Equal(
			t,
			validKey.PubKey.String(),
			vk.PubKey,
		)
	})

	t.Run("validator key address", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		keyPath := filepath.Join(dirPath, defaultValidatorKeyName)

		validKey, err := signer.GeneratePersistedFileKey(keyPath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			dirPath,
			fmt.Sprintf("%s.%s", validatorPrivateKeyKey, "address"),
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		var address string

		require.NoError(t, json.Unmarshal(mockOutput.Bytes(), &address))

		assert.Equal(
			t,
			validKey.Address.String(),
			address,
		)
	})

	t.Run("validator key address, raw", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		keyPath := filepath.Join(dirPath, defaultValidatorKeyName)

		validKey, err := signer.GeneratePersistedFileKey(keyPath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			dirPath,
			fmt.Sprintf("%s.%s", validatorPrivateKeyKey, "address"),
			"--raw",
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		assert.Equal(
			t,
			validKey.Address.String(),
			escapeNewline(mockOutput.Bytes()),
		)
	})

	t.Run("validator key pubkey", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		keyPath := filepath.Join(dirPath, defaultValidatorKeyName)

		validKey, err := signer.GeneratePersistedFileKey(keyPath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			dirPath,
			fmt.Sprintf("%s.%s", validatorPrivateKeyKey, "pub_key"),
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		var address string

		require.NoError(t, json.Unmarshal(mockOutput.Bytes(), &address))

		assert.Equal(
			t,
			validKey.PubKey.String(),
			address,
		)
	})

	t.Run("validator key pubkey, raw", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		keyPath := filepath.Join(dirPath, defaultValidatorKeyName)

		validKey, err := signer.GeneratePersistedFileKey(keyPath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			dirPath,
			fmt.Sprintf("%s.%s", validatorPrivateKeyKey, "pub_key"),
			"--raw",
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		assert.Equal(
			t,
			validKey.PubKey.String(),
			escapeNewline(mockOutput.Bytes()),
		)
	})
}

func TestSecrets_Get_ValidatorStateInfo(t *testing.T) {
	t.Parallel()

	t.Run("validator state info", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		statePath := filepath.Join(dirPath, defaultValidatorStateName)

		validState, err := fstate.GeneratePersistedFileState(statePath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			dirPath,
			validatorStateKey,
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		var vs validatorStateInfo

		require.NoError(t, json.Unmarshal(mockOutput.Bytes(), &vs))

		// Make sure the state info is displayed
		assert.Equal(
			t,
			validState.Step,
			vs.Step,
		)

		assert.Equal(
			t,
			validState.Height,
			vs.Height,
		)

		assert.Equal(
			t,
			validState.Round,
			vs.Round,
		)
	})

	t.Run("validator state info height", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		statePath := filepath.Join(dirPath, defaultValidatorStateName)

		validState, err := fstate.GeneratePersistedFileState(statePath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			dirPath,
			fmt.Sprintf("%s.%s", validatorStateKey, "height"),
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		assert.Equal(
			t,
			fmt.Sprintf("%d\n", validState.Height),
			mockOutput.String(),
		)
	})

	t.Run("validator state info round", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		statePath := filepath.Join(dirPath, defaultValidatorStateName)

		validState, err := fstate.GeneratePersistedFileState(statePath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			dirPath,
			fmt.Sprintf("%s.%s", validatorStateKey, "round"),
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		assert.Equal(
			t,
			fmt.Sprintf("%d\n", validState.Round),
			mockOutput.String(),
		)
	})

	t.Run("validator state info step", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		statePath := filepath.Join(dirPath, defaultValidatorStateName)

		validState, err := fstate.GeneratePersistedFileState(statePath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			dirPath,
			fmt.Sprintf("%s.%s", validatorStateKey, "step"),
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		assert.Equal(
			t,
			fmt.Sprintf("%d\n", validState.Step),
			mockOutput.String(),
		)
	})
}

func TestSecrets_Get_NodeIDInfo(t *testing.T) {
	t.Parallel()

	t.Run("node ID info, default config", func(t *testing.T) {
		t.Parallel()

		cfg := config.DefaultConfig()

		dirPath := t.TempDir()
		nodeKeyPath := filepath.Join(dirPath, defaultNodeKeyName)

		validNodeKey, err := types.GeneratePersistedNodeKey(nodeKeyPath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			dirPath,
			nodeIDKey,
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		var ni nodeIDInfo
		require.NoError(t, json.Unmarshal(mockOutput.Bytes(), &ni))

		// Make sure the node p2p key is displayed
		assert.Equal(
			t,
			validNodeKey.ID().String(),
			ni.ID,
		)

		// Make sure the default node p2p address is displayed
		assert.Equal(
			t,
			constructP2PAddress(validNodeKey.ID(), cfg.P2P.ListenAddress),
			ni.P2PAddress,
		)
	})

	t.Run("node ID info, existing config", func(t *testing.T) {
		t.Parallel()

		var (
			dirPath     = t.TempDir()
			configPath  = constructConfigPath(dirPath)
			secretsPath = constructSecretsPath(dirPath)
			nodeKeyPath = filepath.Join(secretsPath, defaultNodeKeyName)
		)

		// Ensure the sub-dirs exist
		require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0o755))
		require.NoError(t, os.MkdirAll(secretsPath, 0o755))

		// Set up the config
		cfg := config.DefaultConfig()
		cfg.P2P.ListenAddress = "tcp://127.0.0.1:2525"

		require.NoError(t, config.WriteConfigFile(configPath, cfg))

		validNodeKey, err := types.GeneratePersistedNodeKey(nodeKeyPath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			secretsPath,
			nodeIDKey,
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		var ni nodeIDInfo
		require.NoError(t, json.Unmarshal(mockOutput.Bytes(), &ni))

		// Make sure the node p2p key is displayed
		assert.Equal(
			t,
			validNodeKey.ID().String(),
			ni.ID,
		)

		// Make sure the custom node p2p address is displayed
		assert.Equal(
			t,
			constructP2PAddress(validNodeKey.ID(), cfg.P2P.ListenAddress),
			ni.P2PAddress,
		)
	})

	t.Run("ID", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		nodeKeyPath := filepath.Join(dirPath, defaultNodeKeyName)

		validNodeKey, err := types.GeneratePersistedNodeKey(nodeKeyPath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			dirPath,
			fmt.Sprintf("%s.%s", nodeIDKey, "id"),
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		var output string
		require.NoError(t, json.Unmarshal(mockOutput.Bytes(), &output))

		// Make sure the node p2p key is displayed
		assert.Equal(
			t,
			validNodeKey.ID().String(),
			output,
		)
	})

	t.Run("ID, raw", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		nodeKeyPath := filepath.Join(dirPath, defaultNodeKeyName)

		validNodeKey, err := types.GeneratePersistedNodeKey(nodeKeyPath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			dirPath,
			fmt.Sprintf("%s.%s", nodeIDKey, "id"),
			"--raw",
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		// Make sure the node p2p key is displayed
		assert.Equal(
			t,
			validNodeKey.ID().String(),
			escapeNewline(mockOutput.Bytes()),
		)
	})

	t.Run("P2P Address", func(t *testing.T) {
		t.Parallel()

		cfg := config.DefaultConfig()

		dirPath := t.TempDir()
		nodeKeyPath := filepath.Join(dirPath, defaultNodeKeyName)

		validNodeKey, err := types.GeneratePersistedNodeKey(nodeKeyPath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			dirPath,
			fmt.Sprintf("%s.%s", nodeIDKey, "p2p_address"),
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		var output string
		require.NoError(t, json.Unmarshal(mockOutput.Bytes(), &output))

		// Make sure the custom node p2p address is displayed
		assert.Equal(
			t,
			constructP2PAddress(validNodeKey.ID(), cfg.P2P.ListenAddress),
			output,
		)
	})

	t.Run("P2P Address, raw", func(t *testing.T) {
		t.Parallel()

		cfg := config.DefaultConfig()

		dirPath := t.TempDir()
		nodeKeyPath := filepath.Join(dirPath, defaultNodeKeyName)

		validNodeKey, err := types.GeneratePersistedNodeKey(nodeKeyPath)
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--data-dir",
			dirPath,
			fmt.Sprintf("%s.%s", nodeIDKey, "p2p_address"),
			"--raw",
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		// Make sure the custom node p2p address is displayed
		assert.Equal(
			t,
			constructP2PAddress(validNodeKey.ID(), cfg.P2P.ListenAddress),
			escapeNewline(mockOutput.Bytes()),
		)
	})
}
