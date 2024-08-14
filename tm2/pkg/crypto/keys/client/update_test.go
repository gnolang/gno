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

func Test_execUpdate(t *testing.T) {
	t.Parallel()

	// make new test dir
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	defer kbCleanUp()

	// initialize test options
	cfg := &UpdateCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Home:                  kbHome,
				InsecurePasswordStdin: true,
			},
		},
	}

	io := commands.NewTestIO()

	// Add test accounts to keybase.
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	assert.NoError(t, err)

	keyName := "updateApp_Key1"
	p1, p2 := "1234", "foobar"
	mnemonic := "equip will roof matter pink blind book anxiety banner elbow sun young"

	_, err = kb.CreateAccount(keyName, mnemonic, "", p1, 0, 0)
	assert.NoError(t, err)

	{
		// test: Key not found
		args := []string{"blah"}
		io.SetIn(strings.NewReader(p1 + "\n" + p2 + "\n" + p2 + "\n"))
		err = execUpdate(cfg, args, io)
		require.Error(t, err)
		require.Equal(t, "Key blah not found", err.Error())
	}

	{
		// test: Wrong password
		args := []string{keyName}
		io.SetIn(strings.NewReader("blah" + "\n" + p2 + "\n" + p2 + "\n"))
		err = execUpdate(cfg, args, io)
		require.Error(t, err)
		require.Equal(t, "invalid account password", err.Error())
	}

	{
		// test: New passwords don't match
		args := []string{keyName}
		io.SetIn(strings.NewReader(p1 + "\n" + p2 + "\n" + "blah" + "\n"))
		err = execUpdate(cfg, args, io)
		require.Error(t, err)
		require.Equal(t, "unable to parse provided password, passphrases don't match", err.Error())
	}

	{
		// Update the password
		args := []string{keyName}
		io.SetIn(strings.NewReader(p1 + "\n" + p2 + "\n" + p2 + "\n"))
		err = execUpdate(cfg, args, io)
		require.NoError(t, err)
	}

	{
		// test: The old password shouldn't work
		args := []string{keyName}
		io.SetIn(strings.NewReader(p1 + "\n" + p1 + "\n" + p1 + "\n"))
		err = execUpdate(cfg, args, io)
		require.Error(t, err)
		require.Equal(t, "invalid account password", err.Error())
	}

	{
		// Updating the new password to itself should work
		args := []string{keyName}
		io.SetIn(strings.NewReader(p2 + "\n" + p2 + "\n" + p2 + "\n"))
		err = execUpdate(cfg, args, io)
		require.NoError(t, err)
	}
}
