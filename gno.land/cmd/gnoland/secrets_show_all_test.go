package main

import (
	"bytes"
	"context"
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

func TestSecrets_Show_All(t *testing.T) {
	t.Parallel()

	t.Run("invalid data directory", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"show",
			"all",
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
			"all",
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
			"show",
			"all",
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
}
