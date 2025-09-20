package client

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSign_MultisignTx(t *testing.T) {
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
			"multisign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
		}

		assert.ErrorIs(t, cmd.ParseAndRun(ctx, args), flag.ErrHelp)
	})

	t.Run("no signature file provided", func(t *testing.T) {
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
			"multisign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"key-name",
		}

		assert.ErrorIs(t, cmd.ParseAndRun(ctx, args), errNoSignaturesProvided)
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
			"multisign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--signature",
			"./sig.json",
			"TotallyExistingKey",
		}

		assert.True(t, keyerror.IsErrKeyNotFound(cmd.ParseAndRun(ctx, args)))
	})

	t.Run("non-existing multisig reference", func(t *testing.T) {
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
			"multisign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--signature",
			"./sig.json",
			keyName,
		}

		assert.ErrorIs(t, cmd.ParseAndRun(ctx, args), errInvalidMultisigKey)
	})

	t.Run("non-existing tx file", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			key     = secp256k1.GenPrivKey()
			ms      = multisig.NewPubKeyMultisigThreshold(1, []crypto.PubKey{key.PubKey()})
			keyName = "generated-key"
		)

		// Generate a key in the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		_, err = kb.CreateMulti(keyName, ms)
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Create the command
		cmd := NewRootCmdWithBaseConfig(commands.NewTestIO(), baseOptions)

		args := []string{
			"multisign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--tx-path",
			"./TotallyExistingTxFile.json",
			"--signature",
			"./sig.json",
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

			key     = secp256k1.GenPrivKey()
			ms      = multisig.NewPubKeyMultisigThreshold(1, []crypto.PubKey{key.PubKey()})
			keyName = "generated-key"
		)

		// Generate a key in the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		_, err = kb.CreateMulti(keyName, ms)
		require.NoError(t, err)

		// Create an empty tx file
		txFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Create the command
		cmd := NewRootCmdWithBaseConfig(commands.NewTestIO(), baseOptions)

		args := []string{
			"multisign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--tx-path",
			txFile.Name(),
			"--signature",
			"./sig.json",
			keyName,
		}

		assert.ErrorIs(t, cmd.ParseAndRun(ctx, args), errInvalidTxFile)
	})

	t.Run("invalid signature file", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			key     = secp256k1.GenPrivKey()
			ms      = multisig.NewPubKeyMultisigThreshold(1, []crypto.PubKey{key.PubKey()})
			keyName = "generated-key"

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

		info, err := kb.CreateMulti(keyName, ms)
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

		// Create the command
		cmd := NewRootCmdWithBaseConfig(commands.NewTestIO(), baseOptions)

		args := []string{
			"multisign",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--tx-path",
			txFile.Name(),
			"--signature",
			"./sig.json", // invalid
			keyName,
		}

		assert.ErrorIs(t, cmd.ParseAndRun(ctx, args), fs.ErrNotExist)
	})

	t.Run("valid multisig signing", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			encryptPassword = "encrypt"
			multisigName    = "multisig-012"
		)

		// Generate 3 keys, for the multisig
		privKeys := []secp256k1.PrivKeySecp256k1{
			secp256k1.GenPrivKey(),
			secp256k1.GenPrivKey(),
			secp256k1.GenPrivKey(),
		}

		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		// Import the (public) keys into the keybase
		require.NoError(t, kb.ImportPrivKey("k0", privKeys[0], encryptPassword))
		require.NoError(t, kb.ImportPrivKey("k1", privKeys[1], encryptPassword))
		require.NoError(t, kb.ImportPrivKey("k2", privKeys[2], encryptPassword))

		// Build the multisig pub-key (2 of 3)
		msPub := multisig.NewPubKeyMultisigThreshold(
			2, // threshold
			[]crypto.PubKey{
				privKeys[0].PubKey(),
				privKeys[1].PubKey(),
				privKeys[2].PubKey(),
			},
		)

		msInfo, err := kb.CreateMulti(multisigName, msPub)
		require.NoError(t, err)

		// Generate a minimal tx
		tx := std.Tx{
			Fee: std.Fee{
				GasWanted: 10,
				GasFee: std.Coin{
					Amount: 10,
					Denom:  "ugnot",
				},
			},
			Msgs: []std.Msg{
				bank.MsgSend{
					FromAddress: msInfo.GetAddress(), // multisig account is the signer
				},
			},
		}

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		rawTx, err := amino.MarshalJSON(tx)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		// Have 2 out of 3 key sign the tx, with `gnokey sign`
		genSignature := func(keyName, sigOut string) {
			// each invocation needs its own root command
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

			signCmd := NewRootCmdWithBaseConfig(io, baseOptions)

			args := []string{
				"sign",
				"--insecure-password-stdin",
				"--home", kbHome,
				"--tx-path", txFile.Name(),
				"--output-document", sigOut,
				keyName,
			}

			require.NoError(t, signCmd.ParseAndRun(context.Background(), args))
		}

		sigs := []string{
			filepath.Join(t.TempDir(), "sig0.json"),
			filepath.Join(t.TempDir(), "sig1.json"),
		}

		genSignature("k0", sigs[0])
		genSignature("k1", sigs[1])

		// Generate the multisig
		io := commands.NewTestIO()
		multiCmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"multisign",
			"--insecure-password-stdin",
			"--home", kbHome,
			"--tx-path", txFile.Name(),
			"--signature", sigs[0],
			"--signature", sigs[1],
			multisigName,
		}
		require.NoError(t, multiCmd.ParseAndRun(context.Background(), args))

		// Make sure the multisig is valid
		signedRaw, err := os.ReadFile(txFile.Name())
		require.NoError(t, err)

		var signedTx std.Tx
		require.NoError(t, amino.UnmarshalJSON(signedRaw, &signedTx))
		require.Len(t, signedTx.Signatures, 1)

		aggSig := signedTx.Signatures[0]
		require.True(t, aggSig.PubKey.Equals(msPub))

		// Verify the pubkey matches
		require.True(t, msPub.Equals(aggSig.PubKey))

		// Verify the signature
		signBytes, err := signedTx.GetSignBytes("dev", 0, 0)
		assert.NoError(t, err)
		assert.True(t, msPub.VerifyBytes(signBytes, aggSig.Signature))
	})
}
