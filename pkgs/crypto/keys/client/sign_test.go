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

func Test_runSignCmdBasic(t *testing.T) {
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
		DocPath: "",
	}
	cmd.Options = opts

	fakeKeyName1 := "runSignCmd_Key1"
	fakeKeyName2 := "runSignCmd_Key2"
	encPassword := "12345678"

	// add test account to keybase.
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	assert.NoError(t, err)
	_, err = kb.CreateAccount(fakeKeyName1, testMnemonic, "", encPassword, 0, 0)
	assert.NoError(t, err)

	cmd.SetIn(strings.NewReader("XXXDOC"))
	cmd.Args = []string{fakeKeyName1}
	err = runSignCmd(cmd)
	assert.Error(t, err)

	cmd.SetIn(strings.NewReader("XXXDOC\n"))
	cmd.Args = []string{fakeKeyName1}
	err = runSignCmd(cmd)
	assert.Error(t, err)

	cmd.SetIn(strings.NewReader(
		fmt.Sprintf("XXXDOC\n%s\n",
			encPassword,
		),
	))
	cmd.Args = []string{fakeKeyName2}
	err = runSignCmd(cmd)
	assert.Error(t, err)

	cmd.SetIn(strings.NewReader(
		fmt.Sprintf("XXXDOC\n%s\n",
			encPassword,
		),
	))
	cmd.Args = []string{fakeKeyName1}
	err = runSignCmd(cmd)
	assert.NoError(t, err)
}
