package client

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testMnemonic = "equip will roof matter pink blind book anxiety banner elbow sun young"
)

func Test_execDelete(t *testing.T) {
	t.Parallel()

	// make new test dir
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	defer kbCleanUp()

	// initialize test options
	cfg := &DeleteCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Home:                  kbHome,
				InsecurePasswordStdin: true,
			},
		},
	}

	io := commands.NewTestIO()

	fakeKeyName1 := "deleteApp_Key1"
	fakeKeyName2 := "deleteApp_Key2"

	// Add test accounts to keybase.
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	assert.NoError(t, err)

	_, err = kb.CreateAccount(fakeKeyName1, testMnemonic, "", "", 0, 0)
	assert.NoError(t, err)

	_, err = kb.CreateAccount(fakeKeyName2, testMnemonic, "", "", 0, 1)
	assert.NoError(t, err)

	// test: Key not found
	args := []string{"blah"}
	err = execDelete(cfg, args, nil)
	require.Error(t, err)
	require.Equal(t, err.Error(), "Key blah not found")

	// test: User confirmation missing
	args = []string{fakeKeyName1}
	io.SetIn(strings.NewReader(""))
	err = execDelete(cfg, args, io)
	require.Error(t, err)
	require.Equal(t, err.Error(), "EOF")

	{
		_, err = kb.GetByName(fakeKeyName1)
		require.NoError(t, err)

		// Now there is a blank password followed by a confirmation.
		args := []string{fakeKeyName1}
		io.SetIn(strings.NewReader("\ny\n"))
		err = execDelete(cfg, args, io)
		require.NoError(t, err)

		_, err = kb.GetByName(fakeKeyName1)
		require.Error(t, err) // Key1 is gone
	}

	// Set config yes = true
	cfg = &DeleteCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Home:                  kbHome,
				InsecurePasswordStdin: true,
			},
		},
		Yes: true,
	}

	_, err = kb.GetByName(fakeKeyName2)
	require.NoError(t, err)

	// Run again with blank password followed by eof.
	args = []string{fakeKeyName2}
	io.SetIn(strings.NewReader("\n"))
	err = execDelete(cfg, args, io)
	require.NoError(t, err)
	_, err = kb.GetByName(fakeKeyName2)
	require.Error(t, err) // Key2 is gone

	// TODO: Write another case for !keys.Local
}
