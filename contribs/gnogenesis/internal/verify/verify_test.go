package verify

import (
	"context"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/mock"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Verify(t *testing.T) {
	t.Parallel()

	getValidTestGenesis := func() *types.GenesisDoc {
		key := mock.GenPrivKey().PubKey()

		return &types.GenesisDoc{
			GenesisTime:     time.Now(),
			ChainID:         "valid-chain-id",
			ConsensusParams: types.DefaultConsensusParams(),
			Validators: []types.GenesisValidator{
				{
					Address: key.Address(),
					PubKey:  key,
					Power:   1,
					Name:    "valid validator",
				},
			},
			AppState: gnoland.DefaultGenState(),
		}
	}

	t.Run("invalid txs", func(t *testing.T) {
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()

		g.AppState = gnoland.GnoGenesisState{
			Balances: []gnoland.Balance{},
			Txs: []gnoland.TxWithMetadata{
				{},
			},
		}

		require.NoError(t, g.SaveAs(tempFile.Name()))

		// Create the command
		cmd := NewVerifyCmd(commands.NewTestIO())
		args := []string{
			"--genesis-path",
			tempFile.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.Error(t, cmdErr)
	})

	t.Run("invalid tx signature", func(t *testing.T) {
		t.Parallel()

		testTable := []struct {
			name        string
			signBytesFn func(tx *std.Tx, chainID string) []byte
		}{
			{
				name: "invalid chain ID",
				signBytesFn: func(tx *std.Tx, chainID string) []byte {
					// Sign the transaction, but with a chain ID
					// that differs from the genesis chain ID
					signBytes, err := tx.GetSignBytes(chainID+"wrong", 0, 0)
					require.NoError(t, err)

					return signBytes
				},
			},
			{
				name: "invalid account params",
				signBytesFn: func(tx *std.Tx, chainID string) []byte {
					// Sign the transaction, but with an
					// account number that is not 0
					signBytes, err := tx.GetSignBytes(chainID, 10, 0)
					require.NoError(t, err)

					return signBytes
				},
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				tempFile, cleanup := testutils.NewTestFile(t)
				t.Cleanup(cleanup)

				// Generate the genesis
				g := getValidTestGenesis()

				// Generate the transaction
				signer := ed25519.GenPrivKey()

				sendMsg := bank.MsgSend{
					FromAddress: signer.PubKey().Address(),
					ToAddress:   signer.PubKey().Address(),
					Amount:      std.NewCoins(std.NewCoin("ugnot", 10)),
				}

				tx := std.Tx{
					Msgs: []std.Msg{sendMsg},
					Fee: std.Fee{
						GasWanted: 1000000,
						GasFee:    std.NewCoin("ugnot", 20),
					},
				}

				// Sign the transaction
				signBytes := testCase.signBytesFn(&tx, g.ChainID)

				signature, err := signer.Sign(signBytes)
				require.NoError(t, err)

				tx.Signatures = append(tx.Signatures, std.Signature{
					PubKey:    signer.PubKey(),
					Signature: signature,
				})

				appState := g.AppState.(gnoland.GnoGenesisState)
				appState.Txs = []gnoland.TxWithMetadata{
					{
						Tx: tx,
					},
				}
				g.AppState = appState

				require.NoError(t, g.SaveAs(tempFile.Name()))

				// Create the command
				cmd := NewVerifyCmd(commands.NewTestIO())
				args := []string{
					"--genesis-path",
					tempFile.Name(),
				}

				// Run the command
				cmdErr := cmd.ParseAndRun(context.Background(), args)
				assert.ErrorIs(t, cmdErr, errInvalidTxSignature)
			})
		}
	})

	t.Run("invalid balances", func(t *testing.T) {
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()

		g.AppState = gnoland.GnoGenesisState{
			Balances: []gnoland.Balance{
				{},
			},
			Txs: []gnoland.TxWithMetadata{},
		}

		require.NoError(t, g.SaveAs(tempFile.Name()))

		// Create the command
		cmd := NewVerifyCmd(commands.NewTestIO())
		args := []string{
			"--genesis-path",
			tempFile.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.Error(t, cmdErr)
	})

	t.Run("valid genesis", func(t *testing.T) {
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()
		require.NoError(t, g.SaveAs(tempFile.Name()))

		// Create the command
		cmd := NewVerifyCmd(commands.NewTestIO())
		args := []string{
			"--genesis-path",
			tempFile.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)
	})

	t.Run("valid genesis, no state", func(t *testing.T) {
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()
		require.NoError(t, g.SaveAs(tempFile.Name()))

		// Create the command
		cmd := NewVerifyCmd(commands.NewTestIO())
		args := []string{
			"--genesis-path",
			tempFile.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)
	})

	t.Run("hardfork-mode genesis with uncovered validator", func(t *testing.T) {
		// Pre-flight the same invariant gnoland's InitChainer
		// auto-asserts at boot under PastChainIDs (see
		// gno.land/pkg/gnoland/app.go shouldAssertValoperCoverage):
		// every GenesisDoc.Validators entry must have a matching
		// valopers.Register migration tx in state.Txs that derives
		// to the same signing address. Otherwise the chain boots
		// with orphan validators that v3's operator-keyed flow
		// can't manage. Catching this at verify time means the
		// failure surfaces during genesis build, not at first boot.
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()

		// Mark as hardfork-mode: PastChainIDs non-empty triggers the
		// runtime auto-assertion; the verify-time check uses the same
		// gate. With no valopers.Register MsgCalls in state.Txs, the
		// single GenesisDoc.Validators entry is uncovered.
		state := g.AppState.(gnoland.GnoGenesisState)
		state.PastChainIDs = []string{"old-chain"}
		g.AppState = state

		require.NoError(t, g.SaveAs(tempFile.Name()))

		cmd := NewVerifyCmd(commands.NewTestIO())
		args := []string{"--genesis-path", tempFile.Name()}

		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errUncoveredGenesisValidator)
	})

	t.Run("hardfork-mode genesis with covered validator", func(t *testing.T) {
		// Positive case for the same gate: when state.Txs contains a
		// valopers.Register MsgCall whose pubkey argument derives to
		// the validator's address, the coverage check passes.
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()

		// Re-use the genesis validator's PubKey as the Register
		// argument so derive(arg[4]) == Validators[0].Address.
		valPubKey := g.Validators[0].PubKey
		pubKeyBech32 := crypto.PubKeyToBech32(valPubKey)
		opAddr := valPubKey.Address() // any non-zero addr is fine for the verify-time check

		// Sign the Register tx so it passes the existing per-tx
		// signature loop in execVerify before the coverage check runs.
		signer := ed25519.GenPrivKey()
		registerMsg := vm.MsgCall{
			Caller:  signer.PubKey().Address(),
			PkgPath: "gno.land/r/gnops/valopers",
			Func:    "Register",
			Args:    []string{"moul-1", "moul-1's profile", "cloud", opAddr.String(), pubKeyBech32},
		}
		tx := std.Tx{
			Msgs: []std.Msg{registerMsg},
			Fee:  std.Fee{GasWanted: 1000000, GasFee: std.NewCoin("ugnot", 20)},
		}
		signBytes, err := tx.GetSignBytes(g.ChainID, 0, 0)
		require.NoError(t, err)
		signature, err := signer.Sign(signBytes)
		require.NoError(t, err)
		tx.Signatures = append(tx.Signatures, std.Signature{
			PubKey:    signer.PubKey(),
			Signature: signature,
		})

		state := g.AppState.(gnoland.GnoGenesisState)
		state.PastChainIDs = []string{"old-chain"}
		state.Txs = []gnoland.TxWithMetadata{{Tx: tx}}
		g.AppState = state

		require.NoError(t, g.SaveAs(tempFile.Name()))

		cmd := NewVerifyCmd(commands.NewTestIO())
		args := []string{"--genesis-path", tempFile.Name()}

		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)
	})

	t.Run("non-hardfork genesis: coverage check skipped", func(t *testing.T) {
		// Gate symmetry with shouldAssertValoperCoverage: fresh
		// chains (empty PastChainIDs) skip the coverage check at
		// runtime, so verify skips it too. A fresh chain with a
		// validator but no valoper-seed must still verify cleanly.
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis() // PastChainIDs left empty by default
		require.NoError(t, g.SaveAs(tempFile.Name()))

		cmd := NewVerifyCmd(commands.NewTestIO())
		args := []string{"--genesis-path", tempFile.Name()}

		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)
	})

	t.Run("invalid genesis state", func(t *testing.T) {
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()
		g.AppState = "Totally invalid state"
		require.NoError(t, g.SaveAs(tempFile.Name()))

		// Create the command
		cmd := NewVerifyCmd(commands.NewTestIO())
		args := []string{
			"--genesis-path",
			tempFile.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.Error(t, cmdErr)
	})
}
