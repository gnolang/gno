package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	tc := []struct {
		args                 []string
		errShouldContains    string
		stdoutShouldContains string
		stderrShouldContains string
	}{
		// no args
		{[]string{""}, "unknown command", "", ""},
		{[]string{"test"}, "invalid args", "", "Usage: test [test flags] [packages]"},
		{[]string{"build"}, "invalid args", "", "Usage: build [build flags] [packages]"},
		{[]string{"precompile"}, "invalid args", "", "Usage: precompile [precompile flags] [packages]"},
		// {[]string{"repl"}, "", "", ""},

		// --help
		{[]string{"build", "--help"}, "", "# buildOptions options\n-", ""},
		{[]string{"test", "--help"}, "", "# testOptions options\n-", ""},
		{[]string{"precompile", "--help"}, "", "# precompileOptions options\n-", ""},
		{[]string{"repl", "--help"}, "", "# replOptions options\n-", ""},

		// custom
		{[]string{"test", "../../examples/gno.land/p/rand"}, "", "", "ok"},
	}

	for _, test := range tc {
		name := strings.Join(test.args, " ")
		t.Run(name, func(t *testing.T) {
			cmd := command.NewMockCommand()
			mockOut := bytes.NewBufferString("")
			mockErr := bytes.NewBufferString("")
			stdout := command.WriteNopCloser(mockOut)
			stderr := command.WriteNopCloser(mockErr)
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)

			require.NotNil(t, cmd)

			err := runMain(cmd, "gnodev", test.args)

			if test.errShouldContains == "" {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
				require.Contains(t, err.Error(), test.errShouldContains)
			}

			if test.stdoutShouldContains == "" {
				require.Empty(t, mockOut.String())
			} else {
				require.Contains(t, mockOut.String(), test.stdoutShouldContains)
			}

			if test.stderrShouldContains == "" {
				require.Empty(t, mockErr.String())
			} else {
				require.Contains(t, mockErr.String(), test.stderrShouldContains)
			}
		})
	}
}
