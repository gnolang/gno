package client

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/testutils"
	"github.com/jaekwon/testify/assert"
)

func Test_runAddCmdBasic(t *testing.T) {
	cmd := command.NewMockCommand()
	assert.NotNil(t, cmd)

	// make new test dir
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	// initialize test options
	cmd.Options = AddOptions{
		BaseOptions: BaseOptions{
			Home: kbHome,
		},
	}

	cmd.SetIn(strings.NewReader("test1234\ntest1234\n"))
	cmd.Args = []string{"keyname1"}
	err := runAddCmd(cmd)
	assert.NoError(t, err)

	cmd.SetIn(strings.NewReader("test1234\ntest1234\n"))
	cmd.Args = []string{"keyname1"}
	err = runAddCmd(cmd)
	assert.Error(t, err)

	cmd.SetIn(strings.NewReader("y\ntest1234\ntest1234\n"))
	cmd.Args = []string{"keyname1"}
	err = runAddCmd(cmd)
	assert.NoError(t, err)
}
