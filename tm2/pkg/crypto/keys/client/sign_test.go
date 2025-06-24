package client

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestMnemonic generates a random mnemonic
func generateTestMnemonic(t *testing.T) string {
	t.Helper()

	entropy, entropyErr := bip39.NewEntropy(256)
	require.NoError(t, entropyErr)

	mnemonic, mnemonicErr := bip39.NewMnemonic(entropy)
	require.NoError(t, mnemonicErr)

	return mnemonic
}

func TestSign_SignTx(t *testing.T) {
	t.Parallel()

	t.Run("no key provided", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Create the command
		cmd := NewRootCmdWithBaseConfig(commands.NewTestIO(), baseOptions)

		args := []string{
			"sign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
		}

		assert.ErrorIs(t, cmd.ParseAndRun(ctx, args), flag.ErrHelp)
	})

	t.Run("non-existing key", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Create the command
		cmd := NewRootCmdWithBaseConfig(commands.NewTestIO(), baseOptions)

		args := []string{
			"sign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"TotallyExistingKey",
		}

		assert.True(t, keyerror.IsErrKeyNotFound(cmd.ParseAndRun(ctx, args)))
	})

	t.Run("non-existing tx file", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			mnemonic        = generateTestMnemonic(t)
			keyName         = "generated-key"
			encryptPassword = "encrypt"
		)

		// Generate a key in the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		_, err = kb.CreateAccount(keyName, mnemonic, "", encryptPassword, 0, 0)
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Create the command
		cmd := NewRootCmdWithBaseConfig(commands.NewTestIO(), baseOptions)

		args := []string{
			"sign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--tx-path",
			"./TotallyExistingTxFile.json",
			keyName,
		}

		assert.ErrorContains(t, cmd.ParseAndRun(ctx, args), "unable to read transaction file")
	})

	t.Run("empty tx file", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			mnemonic        = generateTestMnemonic(t)
			keyName         = "generated-key"
			encryptPassword = "encrypt"
		)

		// Generate a key in the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		_, err = kb.CreateAccount(keyName, mnemonic, "", encryptPassword, 0, 0)
		require.NoError(t, err)

		// Create an empty tx file
		txFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Create the command
		cmd := NewRootCmdWithBaseConfig(commands.NewTestIO(), baseOptions)

		args := []string{
			"sign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--tx-path",
			txFile.Name(),
			keyName,
		}

		assert.ErrorIs(t, cmd.ParseAndRun(ctx, args), errInvalidTxFile)
	})

	t.Run("corrupted tx amino JSON", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			mnemonic        = generateTestMnemonic(t)
			keyName         = "generated-key"
			encryptPassword = "encrypt"
		)

		// Generate a key in the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		_, err = kb.CreateAccount(keyName, mnemonic, "", encryptPassword, 0, 0)
		require.NoError(t, err)

		// Create an empty tx file
		txFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		// Write invalid JSON
		_, err = txFile.WriteString("{this is absolutely valid JSON]")
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Create the command
		cmd := NewRootCmdWithBaseConfig(commands.NewTestIO(), baseOptions)

		args := []string{
			"sign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--tx-path",
			txFile.Name(),
			keyName,
		}

		assert.ErrorContains(
			t,
			cmd.ParseAndRun(ctx, args),
			"unable to unmarshal transaction",
		)
	})

	t.Run("with output path", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
				Quiet:                 true,
			}

			mnemonic        = generateTestMnemonic(t)
			keyName         = "generated-key"
			encryptPassword = "encrypt"

			tx = std.Tx{
				Fee: std.Fee{
					GasWanted: 10,
					GasFee: std.Coin{
						Amount: 10,
						Denom:  "ugnot",
					},
				},
				Signatures: nil, // no signatures
			}
		)

		// Generate a key in the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		info, err := kb.CreateAccount(keyName, mnemonic, "", encryptPassword, 0, 0)
		require.NoError(t, err)

		// We need to prepare the message signer as well
		// for validation to complete
		tx.Msgs = []std.Msg{
			bank.MsgSend{
				FromAddress: info.GetAddress(),
			},
		}

		// Create an empty tx file
		txFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		// Marshal the tx and write it to the file
		encodedTx, err := amino.MarshalJSON(tx)
		require.NoError(t, err)

		_, err = txFile.Write(encodedTx)
		require.NoError(t, err)

		outputFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Create the command IO
		io := commands.NewTestIO()
		io.SetIn(
			strings.NewReader(
				fmt.Sprintf(
					"%s\n%s\n",
					encryptPassword,
					encryptPassword,
				),
			),
		)

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"sign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--tx-path",
			txFile.Name(),
			"--output-document",
			outputFile.Name(),
			keyName,
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Make sure the output file was updated with the signature
		outputDocumentRaw, err := os.ReadFile(outputFile.Name())
		require.NoError(t, err)

		var sig std.Signature
		require.NoError(t, amino.UnmarshalJSON(outputDocumentRaw, &sig))
		assert.True(t, sig.PubKey.Equals(info.GetPubKey()))

		// Make sure the tx file was not modified
		savedTxRaw, err := os.ReadFile(txFile.Name())
		require.NoError(t, err)

		var savedTx std.Tx
		require.NoError(t, amino.UnmarshalJSON(savedTxRaw, &savedTx))

		require.Len(t, savedTx.Signatures, 0)
	})

	t.Run("invalid tx params", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
				Quiet:                 true,
			}

			mnemonic        = generateTestMnemonic(t)
			keyName         = "generated-key"
			encryptPassword = "encrypt"

			tx = std.Tx{
				Fee: std.Fee{
					GasFee: std.Coin{ // invalid gas fee
						Amount: 0,
						Denom:  "ugnot",
					},
				},
			}
		)

		// Generate a key in the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		_, err = kb.CreateAccount(keyName, mnemonic, "", encryptPassword, 0, 0)
		require.NoError(t, err)

		// Create an empty tx file
		txFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		// Marshal the tx and write it to the file
		encodedTx, err := amino.MarshalJSON(tx)
		require.NoError(t, err)

		_, err = txFile.Write(encodedTx)
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Create the command IO
		io := commands.NewTestIO()
		io.SetIn(
			strings.NewReader(
				fmt.Sprintf(
					"%s\n%s\n",
					encryptPassword,
					encryptPassword,
				),
			),
		)

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"sign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--tx-path",
			txFile.Name(),
			keyName,
		}

		assert.ErrorContains(
			t,
			cmd.ParseAndRun(ctx, args),
			"unable to validate transaction",
		)
	})

	t.Run("empty signature list", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
				Quiet:                 true,
			}

			mnemonic        = generateTestMnemonic(t)
			keyName         = "generated-key"
			encryptPassword = "encrypt"

			tx = std.Tx{
				Fee: std.Fee{
					GasWanted: 10,
					GasFee: std.Coin{
						Amount: 10,
						Denom:  "ugnot",
					},
				},
				Signatures: nil, // no signatures
			}
		)

		// Generate a key in the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		info, err := kb.CreateAccount(keyName, mnemonic, "", encryptPassword, 0, 0)
		require.NoError(t, err)

		// We need to prepare the message signer as well
		// for validation to complete
		tx.Msgs = []std.Msg{
			bank.MsgSend{
				FromAddress: info.GetAddress(),
			},
		}

		// Create an empty tx file
		txFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		// Marshal the tx and write it to the file
		encodedTx, err := amino.MarshalJSON(tx)
		require.NoError(t, err)

		_, err = txFile.Write(encodedTx)
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Create the command IO
		io := commands.NewTestIO()
		io.SetIn(
			strings.NewReader(
				fmt.Sprintf(
					"%s\n%s\n",
					encryptPassword,
					encryptPassword,
				),
			),
		)

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"sign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--tx-path",
			txFile.Name(),
			keyName,
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Make sure the tx file was updated with the signature
		savedTxRaw, err := os.ReadFile(txFile.Name())
		require.NoError(t, err)

		var savedTx std.Tx
		require.NoError(t, amino.UnmarshalJSON(savedTxRaw, &savedTx))

		require.Len(t, savedTx.Signatures, 1)
		assert.True(t, savedTx.Signatures[0].PubKey.Equals(info.GetPubKey()))
	})

	t.Run("existing signature list", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
				Quiet:                 true,
			}

			mnemonic        = generateTestMnemonic(t)
			keyName         = "generated-key"
			encryptPassword = "encrypt"

			anotherKey = "another-key"

			tx = std.Tx{
				Fee: std.Fee{
					GasWanted: 10,
					GasFee: std.Coin{
						Amount: 10,
						Denom:  "ugnot",
					},
				},
			}
		)

		// Generate a key in the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		// Create an initial account
		info, err := kb.CreateAccount(keyName, mnemonic, "", encryptPassword, 0, 0)
		require.NoError(t, err)

		// Create a new account
		anotherKeyInfo, err := kb.CreateAccount(anotherKey, mnemonic, "", encryptPassword, 0, 1)
		require.NoError(t, err)

		// Generate the signature
		signBytes, err := tx.GetSignBytes("id", 1, 0)
		require.NoError(t, err)

		signature, pubKey, err := kb.Sign(anotherKey, encryptPassword, signBytes)
		require.NoError(t, err)

		tx.Signatures = []std.Signature{
			{
				PubKey:    pubKey,
				Signature: signature,
			},
		}

		// We need to prepare the message signers as well
		// for validation to complete
		tx.Msgs = []std.Msg{
			bank.MsgSend{
				FromAddress: info.GetAddress(),
			},
			bank.MsgSend{
				FromAddress: anotherKeyInfo.GetAddress(),
			},
		}

		// Create an empty tx file
		txFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		// Marshal the tx and write it to the file
		encodedTx, err := amino.MarshalJSON(tx)
		require.NoError(t, err)

		_, err = txFile.Write(encodedTx)
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Create the command IO
		io := commands.NewTestIO()
		io.SetIn(
			strings.NewReader(
				fmt.Sprintf(
					"%s\n%s\n",
					encryptPassword,
					encryptPassword,
				),
			),
		)

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"sign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--tx-path",
			txFile.Name(),
			keyName,
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Make sure the tx file was updated with the signature
		savedTxRaw, err := os.ReadFile(txFile.Name())
		require.NoError(t, err)

		var savedTx std.Tx
		require.NoError(t, amino.UnmarshalJSON(savedTxRaw, &savedTx))

		require.Len(t, savedTx.Signatures, 2)
		assert.True(t, savedTx.Signatures[0].PubKey.Equals(anotherKeyInfo.GetPubKey()))
		assert.True(t, savedTx.Signatures[1].PubKey.Equals(info.GetPubKey()))
		assert.NotEqual(t, savedTx.Signatures[0].Signature, savedTx.Signatures[1].Signature)
	})

	t.Run("overwrite existing signature", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
				Quiet:                 true,
			}

			mnemonic        = generateTestMnemonic(t)
			keyName         = "generated-key"
			encryptPassword = "encrypt"

			tx = std.Tx{
				Fee: std.Fee{
					GasWanted: 10,
					GasFee: std.Coin{
						Amount: 10,
						Denom:  "ugnot",
					},
				},
			}
		)

		// Generate a key in the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		info, err := kb.CreateAccount(keyName, mnemonic, "", encryptPassword, 0, 0)
		require.NoError(t, err)

		// Generate the signature
		signBytes, err := tx.GetSignBytes("id", 0, 0)
		require.NoError(t, err)

		signature, pubKey, err := kb.Sign(keyName, encryptPassword, signBytes)
		require.NoError(t, err)

		tx.Signatures = []std.Signature{
			{
				PubKey:    pubKey,
				Signature: signature,
			},
		}

		// We need to prepare the message signer as well
		// for validation to complete
		tx.Msgs = []std.Msg{
			bank.MsgSend{
				FromAddress: info.GetAddress(),
			},
		}

		// Create an empty tx file
		txFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		// Marshal the tx and write it to the file
		encodedTx, err := amino.MarshalJSON(tx)
		require.NoError(t, err)

		_, err = txFile.Write(encodedTx)
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Create the command IO
		io := commands.NewTestIO()
		io.SetIn(
			strings.NewReader(
				fmt.Sprintf(
					"%s\n%s\n",
					encryptPassword,
					encryptPassword,
				),
			),
		)

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"sign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--tx-path",
			txFile.Name(),
			keyName,
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Make sure the tx file was updated with the signature
		savedTxRaw, err := os.ReadFile(txFile.Name())
		require.NoError(t, err)

		var savedTx std.Tx
		require.NoError(t, amino.UnmarshalJSON(savedTxRaw, &savedTx))

		require.Len(t, savedTx.Signatures, 1)
		assert.True(t, savedTx.Signatures[0].PubKey.Equals(info.GetPubKey()))
	})
}
