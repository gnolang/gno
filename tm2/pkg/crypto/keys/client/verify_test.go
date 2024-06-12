package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func Test_verify(t *testing.T) {
	t.Parallel()

	t.Run("verify after signed", func(t *testing.T) {
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
		assert.NoError(t, err)

		info, err := kb.CreateAccount(keyName, mnemonic, "", encryptPassword, 0, 0)
		assert.NoError(t, err)

		// We need to prepare the message signer as well
		// for validation to complete
		tx.Msgs = []std.Msg{
			vm.MsgCall{
				Caller:  info.GetAddress(),
				Send:    std.Coins{},
				PkgPath: "gno.land/r/demo/demo",
				Func:    "Code",
				Args:    []string{},
			},
		}

		// Create an empty tx file
		txFile, err := os.CreateTemp("", "")
		assert.NoError(t, err)

		// Marshal the tx and write it to the file
		encodedTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		_, err = txFile.Write(encodedTx)
		assert.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
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
		assert.NoError(t, cmd.ParseAndRun(ctx, args))

		// Make sure the tx file was updated with the signature
		savedTxRaw, err := os.ReadFile(txFile.Name())
		assert.NoError(t, err)

		var savedTx std.Tx
		assert.NoError(t, amino.UnmarshalJSON(savedTxRaw, &savedTx))

		assert.Len(t, savedTx.Signatures, 1)
		assert.True(t, savedTx.Signatures[0].PubKey.Equals(info.GetPubKey()))

		////////////////////////// VERIFY

		// initialize test options
		sig := savedTx.Signatures[0].Signature
		sigEncoded := hex.EncodeToString(sig)
		argsV := []string{
			"verify",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--docpath",
			txFile.Name(),
			"--chainid",
			"dev",
			"--account-number",
			"0",
			"--account-sequence",
			"0",
			keyName,
			sigEncoded,
		}

		badSig := testutils.MutateByteSlice(sig)
		badSigEncoded := hex.EncodeToString(badSig)
		argBadV := []string{
			"verify",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--docpath",
			txFile.Name(),
			"--chainid",
			"id",
			"--account-number",
			"0",
			"--account-sequence",
			"0",
			keyName,
			badSigEncoded,
		}
		cmdV := NewRootCmdWithBaseConfig(io, baseOptions)
		assert.NoError(t, cmdV.ParseAndRun(ctx, argsV))
		cmdBadV := NewRootCmdWithBaseConfig(io, baseOptions)
		assert.Error(t, cmdBadV.ParseAndRun(ctx, argBadV))
	})

	t.Run("verify empty signature", func(t *testing.T) {
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
		assert.NoError(t, err)

		info, err := kb.CreateAccount(keyName, mnemonic, "", encryptPassword, 0, 0)
		assert.NoError(t, err)

		// We need to prepare the message signer as well
		// for validation to complete
		tx.Msgs = []std.Msg{
			vm.MsgCall{
				Caller:  info.GetAddress(),
				Send:    std.Coins{},
				PkgPath: "gno.land/r/demo/demo",
				Func:    "Code",
				Args:    []string{},
			},
		}

		// Create an empty tx file
		txFile, err := os.CreateTemp("", "")
		assert.NoError(t, err)

		// Marshal the tx and write it to the file
		encodedTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		_, err = txFile.Write(encodedTx)
		assert.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
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
		assert.NoError(t, cmd.ParseAndRun(ctx, args))

		// Make sure the tx file was updated with the signature
		savedTxRaw, err := os.ReadFile(txFile.Name())
		assert.NoError(t, err)

		var savedTx std.Tx
		assert.NoError(t, amino.UnmarshalJSON(savedTxRaw, &savedTx))

		////////////////////////// VERIFY

		// initialize test options
		argsV := []string{
			"verify",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--docpath",
			txFile.Name(),
			"--chainid",
			"id",
			"--account-number",
			"0",
			"--account-sequence",
			"0",
			keyName,
		}
		cmdV := NewRootCmdWithBaseConfig(io, baseOptions)
		assert.Error(t, cmdV.ParseAndRun(ctx, argsV))
	})

	t.Run("verify empty account number", func(t *testing.T) {
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
		assert.NoError(t, err)

		info, err := kb.CreateAccount(keyName, mnemonic, "", encryptPassword, 0, 0)
		assert.NoError(t, err)

		// We need to prepare the message signer as well
		// for validation to complete
		tx.Msgs = []std.Msg{
			vm.MsgCall{
				Caller:  info.GetAddress(),
				Send:    std.Coins{},
				PkgPath: "gno.land/r/demo/demo",
				Func:    "Code",
				Args:    []string{},
			},
		}

		// Create an empty tx file
		txFile, err := os.CreateTemp("", "")
		assert.NoError(t, err)

		// Marshal the tx and write it to the file
		encodedTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		_, err = txFile.Write(encodedTx)
		assert.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
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
		assert.NoError(t, cmd.ParseAndRun(ctx, args))

		// Make sure the tx file was updated with the signature
		savedTxRaw, err := os.ReadFile(txFile.Name())
		assert.NoError(t, err)

		var savedTx std.Tx
		assert.NoError(t, amino.UnmarshalJSON(savedTxRaw, &savedTx))

		////////////////////////// VERIFY

		// initialize test options
		sig := savedTx.Signatures[0].Signature
		sigEncoded := hex.EncodeToString(sig)
		argsV := []string{
			"verify",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--docpath",
			txFile.Name(),
			"--chainid",
			"id",
			"--account-number",
			"0",
			"--account-sequence",
			"0",
			keyName,
			sigEncoded,
		}
		cmdV := NewRootCmdWithBaseConfig(io, baseOptions)
		assert.Error(t, cmdV.ParseAndRun(ctx, argsV))
	})

	t.Run("verify basic", func(t *testing.T) {
		t.Parallel()

		// make new test dir
		kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
		assert.NotNil(t, kbHome)
		defer kbCleanUp()

		io := commands.NewTestIO()

		fakeKeyName1 := "verifyApp_Key1"
		// encPassword := "12345678"
		encPassword := ""
		testMsg := "some message"

		// add test account to keybase.
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		assert.NoError(t, err)
		_, err = kb.CreateAccount(fakeKeyName1, testMnemonic, "", encPassword, 0, 0)
		assert.NoError(t, err)

		// sign test message.
		priv, err := kb.ExportPrivateKeyObject(fakeKeyName1, encPassword)
		assert.NoError(t, err)
		testSig, err := priv.Sign([]byte(testMsg))
		assert.NoError(t, err)
		// testSigHex := hex.EncodeToString(testSig)

		// // good signature passes test.
		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", testMsg),
			),
		)
		err = kb.Verify(fakeKeyName1, []byte(testMsg), testSig)
		assert.NoError(t, err)

		// mutated bad signature fails test.
		testBadSig := testutils.MutateByteSlice(testSig)
		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", testMsg),
			),
		)
		err = kb.Verify(fakeKeyName1, []byte(testMsg), testBadSig)
		assert.Error(t, err)
	})
}
