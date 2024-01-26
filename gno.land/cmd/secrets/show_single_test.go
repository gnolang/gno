package main

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecrets_Show_Single(t *testing.T) {
	t.Parallel()

	t.Run("no individual path set", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"show",
			"single",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errNoOutputSet.Error())
	})

	t.Run("validator key shown", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		keyPath := filepath.Join(dirPath, "key.json")

		validKey := generateValidatorPrivateKey()

		require.NoError(t, saveSecretData(validKey, keyPath))

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"show",
			"single",
			"--validator-key-path",
			keyPath,
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

	t.Run("validator state shown", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		statePath := filepath.Join(dirPath, "state.json")

		validState := generateLastSignValidatorState()

		require.NoError(t, saveSecretData(validState, statePath))

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"show",
			"single",
			"--validator-state-path",
			statePath,
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

	t.Run("node key shown", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		nodeKeyPath := filepath.Join(dirPath, "p2p.json")

		validNodeKey := generateNodeKey()

		require.NoError(t, saveSecretData(validNodeKey, nodeKeyPath))

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"show",
			"single",
			"--node-key-path",
			nodeKeyPath,
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
}
