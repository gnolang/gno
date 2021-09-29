package client

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
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
		TxPath: "",
	}

	fakeKeyName1 := "signApp_Key1"
	fakeKeyName2 := "signApp_Key2"
	encPassword := "12345678"

	// add test account to keybase.
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	assert.NoError(t, err)
	_, err = kb.CreateAccount(fakeKeyName1, testMnemonic, "", encPassword, 0, 0)
	assert.NoError(t, err)

	cmd.SetIn(strings.NewReader("XXXDOC"))
	args := []string{fakeKeyName1}
	err = signApp(cmd, args, opts)
	assert.Error(t, err)

	cmd.SetIn(strings.NewReader("XXXDOC\n"))
	args = []string{fakeKeyName1}
	err = signApp(cmd, args, opts)
	assert.Error(t, err)

	cmd.SetIn(strings.NewReader(
		fmt.Sprintf("XXXDOC\n%s\n",
			encPassword,
		),
	))
	args = []string{fakeKeyName2}
	err = signApp(cmd, args, opts)
	assert.Error(t, err)

	cmd.SetIn(strings.NewReader(
		fmt.Sprintf("XXXDOC\n%s\n",
			encPassword,
		),
	))
	args = []string{fakeKeyName1}
	err = signApp(cmd, args, opts)
	assert.NoError(t, err)
}
