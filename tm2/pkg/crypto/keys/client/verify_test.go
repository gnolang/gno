package client

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
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
