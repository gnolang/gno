package client

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/testutils"
	"github.com/jaekwon/testify/assert"
)

func Test_addAppBasic(t *testing.T) {
	cmd := command.NewMockCommand()
	assert.NotNil(t, cmd)

	// make new test dir
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	// initialize test options
	opts := AddOptions{
		BaseOptions: BaseOptions{
			Home: kbHome,
		},
	}

	cmd.SetIn(strings.NewReader("test1234\ntest1234\n"))
	err := addApp(cmd, []string{"keyname1"}, opts)
	assert.NoError(t, err)

	cmd.SetIn(strings.NewReader("test1234\ntest1234\n"))
	err = addApp(cmd, []string{"keyname1"}, opts)
	assert.Error(t, err)

	cmd.SetIn(strings.NewReader("y\ntest1234\ntest1234\n"))
	err = addApp(cmd, []string{"keyname1"}, opts)
	assert.NoError(t, err)
}
