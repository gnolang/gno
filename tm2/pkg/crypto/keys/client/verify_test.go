package client

import (
	"context"
	"encoding/base64"
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

	// make new test dir
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

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

	signedTx, err := generateSignature(&tx, kb, signOpts, keyOpts)
	assert.NoError(t, err)
	signedTxBase64 := base64.StdEncoding.EncodeToString(signedTx.Signature)

	// Prepare the verification
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

	// Marshal the tx
	encodedTx, err := amino.MarshalJSON(tx)
	assert.NoError(t, err)

	// good signature passes test.
	args := []string{fakeKeyName1, signedTxBase64}
	io.SetIn(
		strings.NewReader(
			fmt.Sprintf("%s\n", encodedTx),
		),
	)
	err = execVerify(cfg, args, io)
	assert.NoError(t, err)

	// mutated bad signature fails test.
	testBadSig := testutils.MutateByteSlice(signedTx.PubKey.Bytes())
	badSignedTxBase64 := base64.StdEncoding.EncodeToString(testBadSig)
	args = []string{fakeKeyName1, badSignedTxBase64}
	io.SetIn(
		strings.NewReader(
			fmt.Sprintf("%s\n", encodedTx),
		),
	)
	err = execVerify(cfg, args, io)
	assert.Error(t, err)
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

	aggSig := signedTx.Signatures[0]
	aggSigB64 := base64.StdEncoding.EncodeToString(aggSig.Signature)

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

	vargs := []string{multisigName, aggSigB64}
	io.SetIn(
		strings.NewReader(
			fmt.Sprintf("%s\n", rawTx),
		),
	)
	err = execVerify(cfg, vargs, io)
	assert.NoError(t, err)
}
