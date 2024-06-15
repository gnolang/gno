package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/p2p"
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
		nodeKey, err := readSecretData[p2p.NodeKey](nodeKeyPath)
		require.NoError(t, err)

		// Get the validator private key
		validatorKeyPath := filepath.Join(tempDir, defaultValidatorKeyName)
		validatorKey, err := readSecretData[privval.FilePVKey](validatorKeyPath)
		require.NoError(t, err)

		// Get the validator state
		validatorStatePath := filepath.Join(tempDir, defaultValidatorStateName)
		state, err := readSecretData[privval.FilePVLastSignState](validatorStatePath)
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

	t.Run("all secrets shown, JSON", func(t *testing.T) {
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
		nodeKey, err := readSecretData[p2p.NodeKey](nodeKeyPath)
		require.NoError(t, err)

		// Get the validator private key
		validatorKeyPath := filepath.Join(tempDir, defaultValidatorKeyName)
		validatorKey, err := readSecretData[privval.FilePVKey](validatorKeyPath)
		require.NoError(t, err)

		// Get the validator state
		validatorStatePath := filepath.Join(tempDir, defaultValidatorStateName)
		state, err := readSecretData[privval.FilePVLastSignState](validatorStatePath)
		require.NoError(t, err)

		// Run the show command
		showArgs := []string{
			"secrets",
			"get",
			"--data-dir",
			tempDir,
			"--json",
		}

		require.NoError(t, cmd.ParseAndRun(context.Background(), showArgs))

		// Set up the output scanner, since JSON output is per-line
		scanner := bufio.NewScanner(bytes.NewReader(mockOutput.Bytes()))

		// Node P2P key //

		require.True(t, scanner.Scan())
		output := scanner.Text()

		var nodeInfo nodeKeyInfo
		require.NoError(t, json.Unmarshal([]byte(output), &nodeInfo))

		assert.Equal(t, nodeInfo.NodeID, nodeKey.ID().String())

		// Validator private key //

		require.True(t, scanner.Scan())
		output = scanner.Text()

		var validatorInfo validatorKeyInfo
		require.NoError(t, json.Unmarshal([]byte(output), &validatorInfo))

		assert.Equal(t, validatorInfo.Address, validatorKey.Address.String())
		assert.Equal(t, validatorInfo.PubKey, validatorKey.PubKey.String())

		// Validator state //

		require.True(t, scanner.Scan())
		output = scanner.Text()

		var validatorState validatorStateInfo
		require.NoError(t, json.Unmarshal([]byte(output), &validatorState))

		assert.Equal(t, validatorState.Step, state.Step)
		assert.Equal(t, validatorState.Height, state.Height)
		assert.Equal(t, validatorState.Round, state.Round)

		require.NoError(t, scanner.Err())
	})
}

func TestSecrets_Get_Single(t *testing.T) {
	t.Parallel()

	t.Run("validator key shown", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		keyPath := filepath.Join(dirPath, defaultValidatorKeyName)

		validKey := generateValidatorPrivateKey()

		require.NoError(t, saveSecretData(validKey, keyPath))

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

		output := mockOutput.String()

		// Make sure the private key info is displayed
		assert.Contains(
			t,
			output,
			validKey.Address.String(),
		)

		assert.Contains(
			t,
			output,
			validKey.PubKey.String(),
		)
	})

	t.Run("validator key shown, JSON", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		keyPath := filepath.Join(dirPath, defaultValidatorKeyName)

		validKey := generateValidatorPrivateKey()

		require.NoError(t, saveSecretData(validKey, keyPath))

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
			"--json",
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		output := mockOutput.String()

		var validatorInfo validatorKeyInfo
		require.NoError(t, json.Unmarshal([]byte(output), &validatorInfo))

		assert.Equal(t, validatorInfo.Address, validKey.Address.String())
		assert.Equal(t, validatorInfo.PubKey, validKey.PubKey.String())
	})

	t.Run("validator state shown", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		statePath := filepath.Join(dirPath, defaultValidatorStateName)

		validState := generateLastSignValidatorState()

		require.NoError(t, saveSecretData(validState, statePath))

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

		output := mockOutput.String()

		// Make sure the state info is displayed
		assert.Contains(
			t,
			output,
			fmt.Sprintf("%d", validState.Step),
		)

		assert.Contains(
			t,
			output,
			fmt.Sprintf("%d", validState.Height),
		)

		assert.Contains(
			t,
			output,
			strconv.Itoa(validState.Round),
		)
	})

	t.Run("validator state shown, JSON", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		statePath := filepath.Join(dirPath, defaultValidatorStateName)

		validState := generateLastSignValidatorState()

		require.NoError(t, saveSecretData(validState, statePath))

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
			"--json",
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		output := mockOutput.String()

		// Make sure the state info is displayed
		var validatorState validatorStateInfo
		require.NoError(t, json.Unmarshal([]byte(output), &validatorState))

		assert.Equal(t, validatorState.Step, validState.Step)
		assert.Equal(t, validatorState.Height, validState.Height)
		assert.Equal(t, validatorState.Round, validState.Round)
	})

	t.Run("node key shown", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		nodeKeyPath := filepath.Join(dirPath, defaultNodeKeyName)

		validNodeKey := generateNodeKey()

		require.NoError(t, saveSecretData(validNodeKey, nodeKeyPath))

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
			nodeKeyKey,
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		output := mockOutput.String()

		// Make sure the node p2p key is displayed
		assert.Contains(
			t,
			output,
			validNodeKey.ID().String(),
		)
	})

	t.Run("node key shown, JSON", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		nodeKeyPath := filepath.Join(dirPath, defaultNodeKeyName)

		validNodeKey := generateNodeKey()

		require.NoError(t, saveSecretData(validNodeKey, nodeKeyPath))

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
			nodeKeyKey,
			"--json",
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		output := mockOutput.String()

		// Make sure the node p2p key is displayed
		var nodeInfo nodeKeyInfo
		require.NoError(t, json.Unmarshal([]byte(output), &nodeInfo))

		assert.Equal(t, nodeInfo.NodeID, validNodeKey.ID().String())
	})
}
