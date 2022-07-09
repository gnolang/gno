package client

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	testutils2 "github.com/gnolang/gno/pkgs/sdk/testutils"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/pkgs/testutils"
	"github.com/jaekwon/testify/assert"
)

func Test_signAppBasic(t *testing.T) {
	cmd := command.NewMockCommand()
	assert.NotNil(t, cmd)

	// make new test dir
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	// initialize test options
	opts := SignOptions{
		BaseOptions: BaseOptions{
			Home: kbHome,
		},
		TxPath:        "-", // stdin
		ChainID:       "testchain",
		AccountNumber: new(uint64),
		Sequence:      new(uint64),
	}

	fakeKeyName1 := "signApp_Key1"
	fakeKeyName2 := "signApp_Key2"
	encPassword := "12345678"

	// add test account to keybase.
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	assert.NoError(t, err)
	acc, err := kb.CreateAccount(fakeKeyName1, testMnemonic, "", encPassword, 0, 0)
	addr := acc.GetAddress()
	assert.NoError(t, err)

	// create a tx to sign.
	msg := testutils2.NewTestMsg(addr)
	fee := std.NewFee(1, std.Coin{"ugnot", 1})
	tx := std.NewTx([]std.Msg{msg}, fee, nil, "")
	txjson := string(amino.MustMarshalJSON(tx))

	cmd.SetIn(strings.NewReader(txjson))
	args := []string{fakeKeyName1}
	err = signApp(cmd, args, opts)
	assert.Error(t, err)

	cmd.SetIn(strings.NewReader(txjson + "\n"))
	args = []string{fakeKeyName1}
	err = signApp(cmd, args, opts)
	assert.Error(t, err)

	cmd.SetIn(strings.NewReader(
		fmt.Sprintf("%s\n%s\n",
			txjson,
			encPassword,
		),
	))
	args = []string{fakeKeyName2}
	err = signApp(cmd, args, opts)
	assert.Error(t, err)

	cmd.SetIn(strings.NewReader(
		fmt.Sprintf("%s\n%s\n",
			txjson,
			encPassword,
		),
	))
	args = []string{fakeKeyName1}
	err = signApp(cmd, args, opts)
	assert.NoError(t, err)
}
