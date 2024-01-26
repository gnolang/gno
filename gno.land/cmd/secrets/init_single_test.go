package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecrets_Init_Single(t *testing.T) {
	t.Parallel()

	t.Run("no individual path set", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"init",
			"single",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errNoOutputSet.Error())
	})

	t.Run("individual secrets initialized", func(t *testing.T) {
		t.Parallel()

		testTable := []struct {
			name      string
			flagValue string
			verifyFn  func(*testing.T, string)
		}{
			{
				"validator key initialized",
				"--validator-key-path",
				verifyValidatorKey,
			},
			{
				"validator state initialized",
				"--validator-state-path",
				verifyValidatorState,
			},
			{
				"node p2p initialized",
				"--node-key-path",
				verifyNodeKey,
			},
		}

		for _, testCase := range testTable {
			testCase := testCase

			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				var (
					tempDir  = t.TempDir()
					dataName = "data.json"

					expectedPath = filepath.Join(tempDir, dataName)
				)

				// Create the command
				cmd := newRootCmd(commands.NewTestIO())
				args := []string{
					"init",
					"single",
					testCase.flagValue,
					expectedPath,
				}

				// Run the command
				cmdErr := cmd.ParseAndRun(context.Background(), args)
				require.NoError(t, cmdErr)

				// Verify the validator key is saved
				testCase.verifyFn(t, expectedPath)
			})
		}
	})
}
