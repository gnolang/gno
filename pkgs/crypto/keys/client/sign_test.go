package client

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/commands"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	sdkutils "github.com/gnolang/gno/pkgs/sdk/testutils"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/pkgs/testutils"
	"github.com/stretchr/testify/assert"
)

func Test_execSign(t *testing.T) {
	t.Parallel()

	// make new test dir
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	// initialize test options
	cfg := &signCfg{
		rootCfg: &baseCfg{
			BaseOptions: BaseOptions{
				Home:                  kbHome,
				InsecurePasswordStdin: true,
			},
		},
		txPath:        "-", // stdin
		chainID:       "dev",
		accountNumber: 0,
		sequence:      0,
	}

	fakeKeyName1 := "signApp_Key1"
	fakeKeyName2 := "signApp_Key2"
	encPassword := "12345678"

	io := commands.NewTestIO()

	// add test account to keybase.
	kb, err := keys.NewKeyBaseFromDir(cfg.rootCfg.Home)
	assert.NoError(t, err)
	acc, err := kb.CreateAccount(fakeKeyName1, testMnemonic, "", encPassword, 0, 0)
	addr := acc.GetAddress()
	assert.NoError(t, err)

	// create a tx to sign.
	msg := sdkutils.NewTestMsg(addr)
	fee := std.NewFee(1, std.Coin{"ugnot", 1000000})
	tx := std.NewTx([]std.Msg{msg}, fee, nil, "")
	txjson := string(amino.MustMarshalJSON(tx))

	args := []string{fakeKeyName1}
	io.SetIn(strings.NewReader(txjson))
	err = execSign(cfg, args, io)
	assert.Error(t, err)

	args = []string{fakeKeyName1}
	io.SetIn(strings.NewReader(txjson + "\n"))
	err = execSign(cfg, args, io)
	assert.Error(t, err)

	args = []string{fakeKeyName2}
	io.SetIn(strings.NewReader(
		fmt.Sprintf("%s\n%s\n",
			txjson,
			encPassword,
		),
	))
	err = execSign(cfg, args, io)
	assert.Error(t, err)

	args = []string{fakeKeyName1}
	io.SetIn(strings.NewReader(
		fmt.Sprintf("%s\n%s\n",
			txjson,
			encPassword,
		),
	))
	err = execSign(cfg, args, io)
	assert.NoError(t, err)
}
