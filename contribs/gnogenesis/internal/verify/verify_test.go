package verify

import (
	"context"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
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
