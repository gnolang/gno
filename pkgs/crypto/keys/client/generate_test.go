package client

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/jaekwon/testify/assert"
	"github.com/jaekwon/testify/require"
)

func Test_RunGenerateCmdNormal(t *testing.T) {
	cmd := command.NewMockCommand()
	assert.NotNil(t, cmd)

	cmd.Options = GenerateOptions{
		CustomEntropy: false,
	}
	cmd.Args = []string{}
	err := runGenerateCmd(cmd)
	require.NoError(t, err)
}

func Test_RunGenerateCmdUser(t *testing.T) {
	cmd := command.NewMockCommand()
	assert.NotNil(t, cmd)

	cmd.Options = GenerateOptions{
		CustomEntropy: true,
	}
	cmd.Args = []string{}
	err := runGenerateCmd(cmd)
	require.Error(t, err)
	require.Equal(t, err.Error(), "EOF")

	// Try again
	cmd.SetIn(strings.NewReader("Hi!\n"))
	cmd.Args = []string{}
	err = runGenerateCmd(cmd)
	require.Error(t, err)
	require.Equal(t, err.Error(),
		"256-bits is 43 characters in Base-64, and 100 in Base-6. You entered 3, and probably want more")

	// Now provide "good" entropy :)
	fakeEntropy := strings.Repeat(":)", 40) + "\ny\n" // entropy + accept count
	cmd.SetIn(strings.NewReader(fakeEntropy))
	cmd.Args = []string{}
	err = runGenerateCmd(cmd)
	require.NoError(t, err)

	// Now provide "good" entropy but no answer
	fakeEntropy = strings.Repeat(":)", 40) + "\n" // entropy + accept count
	cmd.SetIn(strings.NewReader(fakeEntropy))
	cmd.Args = []string{}
	err = runGenerateCmd(cmd)
	require.Error(t, err)

	// Now provide "good" entropy but say no
	fakeEntropy = strings.Repeat(":)", 40) + "\nn\n" // entropy + accept count
	cmd.SetIn(strings.NewReader(fakeEntropy))
	cmd.Args = []string{}
	err = runGenerateCmd(cmd)
	require.NoError(t, err)
}
