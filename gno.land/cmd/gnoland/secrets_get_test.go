package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecrets_Get_All(t *testing.T) {
	t.Parallel()

	t.Run("all secrets shown", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir(), withSecrets)

		secrets, err := homeDir.GetSecrets()
		require.NoError(t, err)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		cmd = newRootCmd(io)

		_ = secrets

		// Get the node key
		nodeKey, err := readSecretData[p2p.NodeKey](homeDir.SecretsNodeKey())
		require.NoError(t, err)

		// Get the validator private key
		validatorKey, err := readSecretData[privval.FilePVKey](homeDir.SecretsValidatorKey())
		require.NoError(t, err)

		// Get the validator state
		state, err := readSecretData[privval.FilePVLastSignState](homeDir.SecretsValidatorState())
		require.NoError(t, err)

		// Run the show command
		showArgs := []string{
			"secrets",
			"get",
			"--home",
			homeDir.Path(),
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

	tests := []struct {
		name     string
		args     []string
		expected func(t *testing.T, out []byte, s *secrets)
	}{
		{
			name: "validator key info",
			args: []string{validatorPrivateKeyKey},
			expected: func(t *testing.T, out []byte, s *secrets) {
				t.Helper()
				var vk validatorKeyInfo

				require.NoError(t, json.Unmarshal(out, &vk))

				// Make sure the private key info is displayed
				assert.Equal(
					t,
					s.ValidatorKeyInfo.Address,
					vk.Address,
				)

				assert.Equal(
					t,
					s.ValidatorKeyInfo.PubKey,
					vk.PubKey,
				)
			},
		},
		{
			name: "validator key address",
			args: []string{fmt.Sprintf("%s.%s", validatorPrivateKeyKey, "address")},
			expected: func(t *testing.T, out []byte, s *secrets) {
				t.Helper()
				var address string

				require.NoError(t, json.Unmarshal(out, &address))

				assert.Equal(
					t,
					s.ValidatorKeyInfo.Address,
					address,
				)
			},
		},
		{
			name: "validator key address, raw",
			args: []string{fmt.Sprintf("%s.%s", validatorPrivateKeyKey, "address"), "--raw"},
			expected: func(t *testing.T, out []byte, s *secrets) {
				t.Helper()
				assert.Equal(
					t,
					s.ValidatorKeyInfo.Address,
					escapeNewline(out),
				)
			},
		},
		{
			name: "validator key pubkey",
			args: []string{fmt.Sprintf("%s.%s", validatorPrivateKeyKey, "pub_key")},
			expected: func(t *testing.T, out []byte, s *secrets) {
				t.Helper()
				var address string

				require.NoError(t, json.Unmarshal(out, &address))

				assert.Equal(
					t,
					s.ValidatorKeyInfo.PubKey,
					address,
				)
			},
		},
		{
			name: "validator key pubkey, raw",
			args: []string{fmt.Sprintf("%s.%s", validatorPrivateKeyKey, "pub_key"), "--raw"},
			expected: func(t *testing.T, out []byte, s *secrets) {
				t.Helper()
				assert.Equal(
					t,
					s.ValidatorKeyInfo.PubKey,
					escapeNewline(out),
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Create a temporary directory
			homeDir := newTestHomeDirectory(t, t.TempDir(), withSecrets)

			secrets, err := homeDir.GetSecrets()
			require.NoError(t, err)

			mockOutput := bytes.NewBufferString("")
			io := commands.NewTestIO()
			io.SetOut(commands.WriteNopCloser(mockOutput))

			// Create the command
			cmd := newRootCmd(io)
			args := []string{
				"secrets",
				"get",
				"--home",
				homeDir.Path(),
			}
			args = append(args, test.args...)

			// Run the command
			require.NoError(t, cmd.ParseAndRun(context.Background(), args))

			test.expected(t, mockOutput.Bytes(), secrets)
		})
	}
}

func TestSecrets_Get_ValidatorStateInfo(t *testing.T) {
	t.Parallel()

	t.Run("validator state info", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir(), withSecrets)

		validState := generateLastSignValidatorState()

		require.NoError(t, saveSecretData(validState, homeDir.SecretsValidatorState()))

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--home",
			homeDir.Path(),
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

	tests := []struct {
		name     string
		key      string
		expected func(t *testing.T, out string, s *secrets)
	}{
		{
			name: "height",
			key:  fmt.Sprintf("%s.%s", validatorStateKey, "height"),
			expected: func(t *testing.T, out string, s *secrets) {
				t.Helper()
				assert.Equal(
					t,
					fmt.Sprintf("%d\n", s.ValidatorStateInfo.Height),
					out,
				)
			},
		},
		{
			name: "round",
			key:  fmt.Sprintf("%s.%s", validatorStateKey, "round"),
			expected: func(t *testing.T, out string, s *secrets) {
				t.Helper()
				assert.Equal(
					t,
					fmt.Sprintf("%d\n", s.ValidatorStateInfo.Round),
					out,
				)
			},
		},
		{
			name: "step",
			key:  fmt.Sprintf("%s.%s", validatorStateKey, "step"),
			expected: func(t *testing.T, out string, s *secrets) {
				t.Helper()
				assert.Equal(
					t,
					fmt.Sprintf("%d\n", s.ValidatorStateInfo.Step),
					out,
				)
			},
		},
	}
	for _, test := range tests {
		t.Run("validator state info "+test.name, func(t *testing.T) {
			t.Parallel()

			// Create a temporary directory
			homeDir := newTestHomeDirectory(t, t.TempDir(), withSecrets)

			secrets, err := homeDir.GetSecrets()
			require.NoError(t, err)

			mockOutput := bytes.NewBufferString("")
			io := commands.NewTestIO()
			io.SetOut(commands.WriteNopCloser(mockOutput))

			// Create the command
			cmd := newRootCmd(io)
			args := []string{
				"secrets",
				"get",
				"--home",
				homeDir.Path(),
				test.key,
			}

			// Run the command
			require.NoError(t, cmd.ParseAndRun(context.Background(), args))

			test.expected(t, mockOutput.String(), secrets)
		})
	}
}

func TestSecrets_Get_NodeIDInfo(t *testing.T) {
	t.Parallel()

	t.Run("node ID info, default config", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir(), withSecrets)

		secrets, err := homeDir.GetSecrets()
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--home",
			homeDir.Path(),
			nodeIDKey,
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		var ni nodeIDInfo
		require.NoError(t, json.Unmarshal(mockOutput.Bytes(), &ni))

		// Make sure the node p2p key is displayed
		assert.Equal(
			t,
			secrets.NodeIDInfo.ID,
			ni.ID,
		)

		// Make sure the default node p2p address is displayed
		assert.Equal(
			t,
			secrets.NodeIDInfo.P2PAddress,
			ni.P2PAddress,
		)
	})

	t.Run("node ID info, existing config", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir(), withConfig, withSecrets)

		cfg, err := homeDir.GetConfig()
		require.NoError(t, err)

		cfg.P2P.ListenAddress = "tcp://127.0.0.1:2525"
		require.NoError(t, config.WriteConfigFile(homeDir.ConfigFile(), cfg))

		validNodeKey := generateNodeKey()
		require.NoError(t, saveSecretData(validNodeKey, homeDir.SecretsNodeKey()))

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--home",
			homeDir.Path(),
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

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir(), withSecrets)

		secrets, err := homeDir.GetSecrets()
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--home",
			homeDir.Path(),
			nodeIDKey,
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		// TODO(albttx): why string not working here
		var output map[string]string
		require.NoError(t, json.Unmarshal(mockOutput.Bytes(), &output))

		// Make sure the node p2p key is displayed
		assert.Equal(
			t,
			secrets.NodeIDInfo.ID,
			output["id"],
		)
	})

	t.Run("ID, raw", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir(), withSecrets)

		secrets, err := homeDir.GetSecrets()
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--home",
			homeDir.Path(),
			fmt.Sprintf("%s.%s", nodeIDKey, "id"),
			"--raw",
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		// Make sure the node p2p key is displayed
		assert.Equal(
			t,
			secrets.NodeIDInfo.ID,
			escapeNewline(mockOutput.Bytes()),
		)
	})

	t.Run("P2P Address", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir(), withSecrets, withConfig)

		secrets, err := homeDir.GetSecrets()
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--home",
			homeDir.Path(),
			fmt.Sprintf("%s.%s", nodeIDKey, "p2p_address"),
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		var output string
		require.NoError(t, json.Unmarshal(mockOutput.Bytes(), &output))

		// Make sure the custom node p2p address is displayed
		assert.Equal(
			t,
			secrets.NodeIDInfo.P2PAddress,
			output,
		)
	})

	t.Run("P2P Address, raw", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir(), withSecrets, withConfig)

		secrets, err := homeDir.GetSecrets()
		require.NoError(t, err)

		mockOutput := bytes.NewBufferString("")
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOutput))

		// Create the command
		cmd := newRootCmd(io)
		args := []string{
			"secrets",
			"get",
			"--home",
			homeDir.Path(),
			fmt.Sprintf("%s.%s", nodeIDKey, "p2p_address"),
			"--raw",
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		// Make sure the custom node p2p address is displayed
		assert.Equal(
			t,
			secrets.NodeIDInfo.P2PAddress,
			// constructP2PAddress(validNodeKey.ID(), cfg.P2P.ListenAddress),
			escapeNewline(mockOutput.Bytes()),
		)
	})
}
