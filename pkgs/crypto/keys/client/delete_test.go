package client

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/testutils"
	"github.com/jaekwon/testify/assert"
	"github.com/jaekwon/testify/require"
)

func Test_runDeleteCmd(t *testing.T) {
	cmd := command.NewMockCommand()
	assert.NotNil(t, cmd)

	// make new test dir
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	defer kbCleanUp()

	// initialize test options
	opts := DeleteOptions{
		BaseOptions: BaseOptions{
			Home: kbHome,
		},
	}
	cmd.Options = opts

	fakeKeyName1 := "runDeleteCmd_Key1"
	fakeKeyName2 := "runDeleteCmd_Key2"

	// Add test accounts to keybase.
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	assert.NoError(t, err)
	_, err = kb.CreateAccount(fakeKeyName1, testMnemonic, "", "", 0, 0)
	assert.NoError(t, err)
	_, err = kb.CreateAccount(fakeKeyName2, testMnemonic, "", "", 0, 1)
	assert.NoError(t, err)

	// test: Key not found
	cmd.Args = []string{"blah"}
	err = runDeleteCmd(cmd)
	require.Error(t, err)
	require.Equal(t, err.Error(), "Key blah not found")

	// test: User confirmation missing
	cmd.Args = []string{fakeKeyName1}
	err = runDeleteCmd(cmd)
	require.Error(t, err)
	require.Equal(t, err.Error(), "EOF")

	{
		_, err = kb.Get(fakeKeyName1)
		require.NoError(t, err)

		// Now there is a blank password followed by a confirmation.
		cmd.SetIn(strings.NewReader("\ny\n"))
		cmd.Args = []string{fakeKeyName1}
		err = runDeleteCmd(cmd)
		require.NoError(t, err)

		_, err = kb.Get(fakeKeyName1)
		require.Error(t, err) // Key1 is gone
	}

	// Set DeleteOptions.Yes = true
	cmd.Options = DeleteOptions{
		BaseOptions: BaseOptions{
			Home: kbHome,
		},
		Yes: true,
	}

	_, err = kb.Get(fakeKeyName2)
	require.NoError(t, err)

	// Run again with blank password followed by eof.
	cmd.SetIn(strings.NewReader("\n"))
	cmd.Args = []string{fakeKeyName2}
	err = runDeleteCmd(cmd)
	require.NoError(t, err)
	_, err = kb.Get(fakeKeyName2)
	require.Error(t, err) // Key2 is gone

	// TODO: Write another case for !keys.Local
}
