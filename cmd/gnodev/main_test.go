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
		args                []string
		errShouldContain    string
		stdoutShouldContain string
		stderrShouldContain string
		panicShouldContain  string
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
		{[]string{"test", "../../tests/integ/minimalist-gno1"}, "", "", "no test files", ""},
		{[]string{"test", "../../tests/integ/minimalist-gno2"}, "", "", "ok", ""},
		{[]string{"test", "../../tests/integ/minimalist-gno3"}, "", "", "ok", ""},
		{[]string{"test", "../../tests/integ/valid1", "--verbose"}, "", "", "ok", ""},
		{[]string{"test", "../../tests/integ/valid2"}, "", "", "ok", ""},
		{[]string{"test", "../../tests/integ/valid2", "--verbose"}, "", "", "ok", ""},
		{[]string{"test", "../../tests/integ/failing1", "--verbose"}, "FAIL: 1 go test errors", "", "FAIL: TestAlwaysFailing", ""},
		{[]string{"test", "../../tests/integ/failing2", "--verbose"}, "", "", "=== RUN", "got unexpected error: beep boop"}, // FIXME: should fail
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

			checkOutputs := func(t *testing.T) {
				t.Helper()

				if test.stdoutShouldContain == "" {
					require.Empty(t, mockOut.String(), "stdout should be empty")
				} else {
					t.Log("out", mockOut.String())
					require.Contains(t, mockOut.String(), test.stdoutShouldContain, "stdout should contain")
				}

				if test.stderrShouldContain == "" {
					require.Empty(t, mockErr.String(), "stderr should be empty")
				} else {
					t.Log("err", mockErr.String())
					require.Contains(t, mockErr.String(), test.stderrShouldContain, "stderr should contain")
				}
			}

			exec := "gnodev"
			defer func() {
				if r := recover(); r != nil {
					require.NotEmpty(t, test.panicShouldContain, "should not panic")
					require.Empty(t, test.errShouldContain, "should not expect an error")
					output := fmt.Sprintf("%v", r)
					t.Log("recover", output)
					require.Contains(t, output, test.panicShouldContain, "recover")
					checkOutputs(t)
				} else {
					require.Empty(t, test.panicShouldContain, "should panic")
				}
			}()
			err := runMain(cmd, exec, test.args)

			if test.errShouldContain == "" {
				require.Nil(t, err, "err should be nil")
			} else {
				t.Log("err", err.Error())
				require.NotNil(t, err, "err shouldn't be nil")
				require.Contains(t, err.Error(), test.errShouldContain, "err should contain")
			}

			checkOutputs(t)
		})
	}
}
