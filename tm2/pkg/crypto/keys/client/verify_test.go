package client

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_execVerify(t *testing.T) {
	t.Parallel()

	const (
		accountNumber   = 10
		accountSequence = 2
		fakeKeyName1    = "verifyApp_Key1"
		encPassword     = ""
		chainID         = "dev"
	)

	prepare := func(t *testing.T) (string, std.Tx, func()) {
		t.Helper()

		// make new test dir
		kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
		assert.NotNil(t, kbHome)

		// add test account to keybase.
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		assert.NoError(t, err)
		info, err := kb.CreateAccount(fakeKeyName1, testMnemonic, "", encPassword, 0, 0)
		assert.NoError(t, err)

		// Prepare the signature
		signOpts := signOpts{
			chainID:         chainID,
			accountSequence: accountSequence,
			accountNumber:   accountNumber,
		}

		keyOpts := keyOpts{
			keyName:     fakeKeyName1,
			decryptPass: "",
		}

		// construct msg & tx and marshal.
		msg := bank.MsgSend{
			FromAddress: info.GetAddress(),
			ToAddress:   info.GetAddress(),
			Amount: std.Coins{
				std.Coin{
					Denom:  "ugnot",
					Amount: 10,
				},
			},
		}

		tx := std.Tx{
			Msgs: []std.Msg{msg},
			Fee: std.Fee{
				GasWanted: 10,
				GasFee: std.Coin{
					Amount: 10,
					Denom:  "ugnot",
				},
			},
		}

		sig, err := generateSignature(&tx, kb, signOpts, keyOpts)
		assert.NoError(t, err)

		// Add signature to the transaction
		tx.Signatures = []std.Signature{*sig}

		return kbHome, tx, kbCleanUp
	}

	t.Run("test number of argument", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{} // NO ARGUMENTS

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test: bad key name", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{"bad-key-name"} // BAD KEY NAME

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test stdin: bad transaction", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)
		// mutate the raw tx to make it bad
		rawTx[0] = 0xFF

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx), // BAD RAW TX
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test stdin: bad newline", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(strings.NewReader(string(rawTx)))

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test stdin: signature ok", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.NoError(t, err)
	})

	t.Run("test stdin: missing signature", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Remove any signatures from the tx
		tx.Signatures = nil

		// Marshal the tx
		rawTxWithoutSig, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		// no signature in tx and no -signature or -sigpath flag
		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTxWithoutSig),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test stdin: -signature flag: bad format", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		sigHex := hex.EncodeToString(tx.Signatures[0].Signature)

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			Signature:       sigHex,
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test stdin: -signature flag: ok", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		sigb64 := base64.StdEncoding.EncodeToString(tx.Signatures[0].Signature)

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			Signature:       sigb64,
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.NoError(t, err)
	})

	t.Run("test stdin: -signature flag: bad signature", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		// mutated bad signature fails test.
		testBadSig := testutils.MutateByteSlice(tx.Signatures[0].PubKey.Bytes())
		badSigb64 := base64.StdEncoding.EncodeToString(testBadSig)

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			Signature:       badSigb64, // BAD SIGNATURE
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test stdin: -sigpath flag: no signature", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		// rawTxWithSig in std.Tx, not std.Signature
		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			SigPath:         txFile.Name(),
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test stdint: -sigpath flag: ok", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		// Marshal the signature
		rawSig, err := amino.MarshalJSON(tx.Signatures[0])
		assert.NoError(t, err)

		sigFile, err := os.CreateTemp("", "sig-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(sigFile.Name(), rawSig, 0o644))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			SigPath:         sigFile.Name(),
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.NoError(t, err)
	})

	t.Run("test stdin: bad -account-sequence flag", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence + 1, // BAD NUMBER
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test stdin: bad -account-number flag", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			AccountNumber:   accountNumber + 1, // BAD NUMBER
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test stdin: bad -chain-id flag", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         "bad-chain-id", // BAD CHAIN ID
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test stdin: try to query network: bad verification", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
					Remote:                "http://localhost:26657", // needs remote to fetch account info
				},
			},
			DocPath: "",
			ChainID: chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io) // account-number and account-sequence wrong
		assert.Error(t, err)
	})

	t.Run("test stdin: no -account-sequence and -account-number flags and -offline: error", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath: "",
			ChainID: "bad-chain-id", // BAD CHAIN ID
			Offline: true,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test: -docpath flag: ok", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         txFile.Name(),
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.NoError(t, err)
	})

	t.Run("test: -docpath flag: bad transaction", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)
		// mutate the raw tx to make it bad
		badRawTx := testutils.MutateByteSlice(rawTx)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), badRawTx, 0o644))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         txFile.Name(),
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	// both -sigpath and -signature flags cannot be used at the same time
	t.Run("test: -sigpath and -signature flags error", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		// Marshal the signature
		rawSig, err := amino.MarshalJSON(tx.Signatures[0])
		assert.NoError(t, err)

		sigFile, err := os.CreateTemp("", "sig-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(sigFile.Name(), rawSig, 0o644))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			DocPath:         "",
			SigPath:         sigFile.Name(), // both flags used
			Signature:       string(rawSig), // both flags used
			AccountNumber:   accountNumber,
			AccountSequence: accountSequence,
			ChainID:         chainID,
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		io.SetIn(
			strings.NewReader(
				fmt.Sprintf("%s\n", rawTx),
			),
		)

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})
}

func Test_VerifyMultisig(t *testing.T) {
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

	// Get the multisig from the transaction file
	signedRaw, err := os.ReadFile(txFile.Name())
	require.NoError(t, err)

	var signedTx std.Tx
	require.NoError(t, amino.UnmarshalJSON(signedRaw, &signedTx))
	require.Len(t, signedTx.Signatures, 1)

	// Prepare the verify function
	cfg := &VerifyCfg{
		RootCfg: &BaseCfg{
			BaseOptions: baseOptions,
		},
		DocPath:         "",
		ChainID:         "dev",
		AccountNumber:   0,
		AccountSequence: 0,
	}

	vargs := []string{multisigName}
	io.SetIn(
		strings.NewReader(
			fmt.Sprintf("%s\n", signedRaw),
		),
	)
	err = execVerify(context.Background(), cfg, vargs, io)
	assert.NoError(t, err)
}
