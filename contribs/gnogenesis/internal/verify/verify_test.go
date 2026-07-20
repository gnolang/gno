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

	t.Run("skip signature check", func(t *testing.T) {
		// Genesis-mode txs can carry signatures that intentionally don't
		// verify: a caller_override patches the caller post-sign (e.g. a
		// names.Enable admin call), and valoper-seed emits zero-value
		// placeholder signatures. Nodes accept both under
		// --skip-genesis-sig-verification; -skip-signature-check is the
		// verify-time equivalent, keeping every other check active.
		t.Parallel()

		testTable := []struct {
			name         string
			signaturesFn func(tx *std.Tx, chainID string) []std.Signature
		}{
			{
				name: "mutated tx body (caller_override)",
				signaturesFn: func(tx *std.Tx, chainID string) []std.Signature {
					// Sign with a chain ID that differs from the genesis
					// chain ID — same mismatch class as a post-sign
					// caller patch: valid signature shape, wrong payload.
					signer := ed25519.GenPrivKey()
					signBytes, err := tx.GetSignBytes(chainID+"wrong", 0, 0)
					require.NoError(t, err)
					signature, err := signer.Sign(signBytes)
					require.NoError(t, err)

					return []std.Signature{
						{
							PubKey:    signer.PubKey(),
							Signature: signature,
						},
					}
				},
			},
			{
				name: "zero-value placeholder signature (valoper-seed)",
				signaturesFn: func(tx *std.Tx, _ string) []std.Signature {
					return make([]std.Signature, len(tx.GetSigners()))
				},
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				tempFile, cleanup := testutils.NewTestFile(t)
				t.Cleanup(cleanup)

				g := getValidTestGenesis()

				sender := ed25519.GenPrivKey()
				sendMsg := bank.MsgSend{
					FromAddress: sender.PubKey().Address(),
					ToAddress:   sender.PubKey().Address(),
					Amount:      std.NewCoins(std.NewCoin("ugnot", 10)),
				}

				tx := std.Tx{
					Msgs: []std.Msg{sendMsg},
					Fee: std.Fee{
						GasWanted: 1000000,
						GasFee:    std.NewCoin("ugnot", 20),
					},
				}
				tx.Signatures = testCase.signaturesFn(&tx, g.ChainID)

				appState := g.AppState.(gnoland.GnoGenesisState)
				appState.Txs = []gnoland.TxWithMetadata{
					{
						Tx: tx,
					},
				}
				g.AppState = appState

				require.NoError(t, g.SaveAs(tempFile.Name()))

				// Without the flag, verification must fail.
				cmd := NewVerifyCmd(commands.NewTestIO())
				args := []string{
					"--genesis-path",
					tempFile.Name(),
				}
				require.Error(t, cmd.ParseAndRun(context.Background(), args))

				// With the flag, every non-signature check still runs
				// and the genesis is accepted.
				cmd = NewVerifyCmd(commands.NewTestIO())
				args = []string{
					"--genesis-path",
					tempFile.Name(),
					"--skip-signature-check",
				}
				require.NoError(t, cmd.ParseAndRun(context.Background(), args))
			})
		}
	})

	t.Run("missing signer public key", func(t *testing.T) {
		// Zero-value placeholder signatures (e.g. valoper-seed output)
		// carry no public key. Verification must reject them with a
		// proper error, not dereference the nil PubKey.
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()

		sender := ed25519.GenPrivKey()
		sendMsg := bank.MsgSend{
			FromAddress: sender.PubKey().Address(),
			ToAddress:   sender.PubKey().Address(),
			Amount:      std.NewCoins(std.NewCoin("ugnot", 10)),
		}

		tx := std.Tx{
			Msgs: []std.Msg{sendMsg},
			Fee: std.Fee{
				GasWanted: 1000000,
				GasFee:    std.NewCoin("ugnot", 20),
			},
		}
		// One zero-value signature per signer: passes ValidateBasic's
		// len(Signatures) == len(GetSigners()) requirement, carries no
		// key material.
		tx.Signatures = make([]std.Signature, len(tx.GetSigners()))

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
		// verify ignores Args[3] (operator addr); only Args[4]
		// (signing pubkey) matters for the coverage check. Runtime
		// distinguishes the two, but any non-zero addr is fine in
		// this verify-time test.
		opAddr := valPubKey.Address()

		// Sign the Register tx so it passes the existing per-tx
		// signature loop in execVerify before the coverage check runs.
		signer := ed25519.GenPrivKey()
		registerMsg := vm.MsgCall{
			Caller:  signer.PubKey().Address(),
			PkgPath: valopersPkgPath,
			Func:    valopersRegisterFn,
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

	t.Run("hardfork-mode genesis with two validators, only one covered", func(t *testing.T) {
		// Exercises the partial-coverage path: the uncovered []string
		// aggregation must list ONLY the uncovered validator, and the
		// error message must surface that address (not the covered one).
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()

		// Add a second validator with a distinct pubkey.
		secondKey := mock.GenPrivKey().PubKey()
		g.Validators = append(g.Validators, types.GenesisValidator{
			Address: secondKey.Address(),
			PubKey:  secondKey,
			Power:   1,
			Name:    "uncovered validator",
		})

		// Build a Register tx that covers ONLY the first validator.
		coveredPubKey := g.Validators[0].PubKey
		pubKeyBech32 := crypto.PubKeyToBech32(coveredPubKey)
		opAddr := coveredPubKey.Address()

		signer := ed25519.GenPrivKey()
		registerMsg := vm.MsgCall{
			Caller:  signer.PubKey().Address(),
			PkgPath: valopersPkgPath,
			Func:    valopersRegisterFn,
			Args:    []string{"covered-1", "covered profile", "cloud", opAddr.String(), pubKeyBech32},
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
		require.ErrorIs(t, cmdErr, errUncoveredGenesisValidator)
		// Only the second (uncovered) validator's address must appear
		// in the error; the first (covered) must not.
		assert.Contains(t, cmdErr.Error(), secondKey.Address().String())
		assert.NotContains(t, cmdErr.Error(), coveredPubKey.Address().String())
	})

	t.Run("hardfork-mode genesis: Failed Register tx does not cover validator", func(t *testing.T) {
		// Runtime gno.land/pkg/gnoland/app.go:779-787 short-circuits
		// metadata.Failed=true txs before baseApp.Deliver runs, so a
		// Failed Register tx never actually populates valoperCache.
		// The verify-time check must skip Failed txs the same way,
		// otherwise verify reports coverage that the runtime ignores —
		// verify-OK followed by chain-PANIC at boot, the exact
		// leaky-shift-left this PR is meant to close.
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()

		valPubKey := g.Validators[0].PubKey
		pubKeyBech32 := crypto.PubKeyToBech32(valPubKey)
		opAddr := valPubKey.Address()

		signer := ed25519.GenPrivKey()
		registerMsg := vm.MsgCall{
			Caller:  signer.PubKey().Address(),
			PkgPath: valopersPkgPath,
			Func:    valopersRegisterFn,
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
		// Failed=true is the only diff vs the "covered" subtest — the
		// asymmetry under test.
		state.Txs = []gnoland.TxWithMetadata{{
			Tx:       tx,
			Metadata: &gnoland.GnoTxMetadata{Failed: true},
		}}
		g.AppState = state

		require.NoError(t, g.SaveAs(tempFile.Name()))

		cmd := NewVerifyCmd(commands.NewTestIO())
		args := []string{"--genesis-path", tempFile.Name()}

		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errUncoveredGenesisValidator)
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
