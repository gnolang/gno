package main

import (
	"bytes"
	"fmt"
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
		panicShouldContains  string
	}{
		// no args
		{[]string{""}, "unknown command", "", "", ""},
		{[]string{"test"}, "invalid args", "", "Usage: test [test flags] [packages]", ""},
		{[]string{"build"}, "invalid args", "", "Usage: build [build flags] [packages]", ""},
		{[]string{"precompile"}, "invalid args", "", "Usage: precompile [precompile flags] [packages]", ""},
		// {[]string{"repl"}, "", "", ""},

		// --help
		{[]string{"build", "--help"}, "", "# buildOptions options\n-", "", ""},
		{[]string{"test", "--help"}, "", "# testOptions options\n-", "", ""},
		{[]string{"precompile", "--help"}, "", "# precompileOptions options\n-", "", ""},
		{[]string{"repl", "--help"}, "", "# replOptions options\n-", "", ""},

		// custom
		{[]string{"test", "../../examples/gno.land/p/rand"}, "", "", "ok", ""},
		{[]string{"test", "../../tests/integ/no-such-dir"}, "no such file or directory", "", "", ""},
		{[]string{"test", "../../tests/integ/empty-dir"}, "", "", "", ""},
		{[]string{"test", "../../tests/integ/empty-gno1"}, "", "", "no test files", ""},
		{[]string{"test", "../../tests/integ/empty-gno2"}, "", "", "", "expected 'package', found 'EOF'"}, // FIXME: better error handling
		{[]string{"test", "../../tests/integ/empty-gno3"}, "", "", "", "expected 'package', found 'EOF'"}, // FIXME: better error handling
	}

	for _, test := range tc {
		name := strings.Join(test.args, " ")
		name = strings.ReplaceAll(name, "/", "~")
		t.Run(name, func(t *testing.T) {
			cmd := command.NewMockCommand()
			mockOut := bytes.NewBufferString("")
			mockErr := bytes.NewBufferString("")
			stdout := command.WriteNopCloser(mockOut)
			stderr := command.WriteNopCloser(mockErr)
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)

			require.NotNil(t, cmd)

			exec := "gnodev"
			defer func() {
				if r := recover(); r != nil {
					output := fmt.Sprintf("%v", r)
					require.Contains(t, output, test.panicShouldContains)
				} else {
					require.Empty(t, test.panicShouldContains)
				}
			}()
			err := runMain(cmd, exec, test.args)

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
