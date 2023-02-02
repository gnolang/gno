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
		args []string

		// for the following FooContain+FooBe expected couples, if both are empty,
		// then the test suite will require that the "got" is not empty.
		errShouldContain     string
		errShouldBe          string
		stderrShouldContain  string
		stdoutShouldBe       string
		stdoutShouldContain  string
		stderrShouldBe       string
		recoverShouldContain string
		recoverShouldBe      string
	}{
		// no args
		{args: []string{""}, errShouldBe: "unknown command "},
		{args: []string{"test"}, errShouldBe: "invalid args", stderrShouldBe: "Usage: test [test flags] [packages]\n"},
		{args: []string{"build"}, errShouldBe: "invalid args", stderrShouldBe: "Usage: build [build flags] [packages]\n"},
		{args: []string{"precompile"}, errShouldBe: "invalid args", stderrShouldBe: "Usage: precompile [precompile flags] [packages]\n"},
		// {args: []string{"repl"}},

		// --help
		{args: []string{"build", "--help"}, stdoutShouldContain: "# buildOptions options\n-"},
		{args: []string{"test", "--help"}, stdoutShouldContain: "# testOptions options\n-"},
		{args: []string{"precompile", "--help"}, stdoutShouldContain: "# precompileFlags options\n-"},
		{args: []string{"repl", "--help"}, stdoutShouldContain: "# replOptions options\n-"},

		// custom
		{args: []string{"test", "../../examples/gno.land/p/demo/rand"}, stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/rand \t"},
		{args: []string{"test", "../../tests/integ/no-such-dir"}, errShouldContain: "no such file or directory"},
		{args: []string{"test", "../../tests/integ/empty-dir"}}, // FIXME: should have an output
		{args: []string{"test", "../../tests/integ/minimalist-gno1"}, stderrShouldBe: "?       ./../../tests/integ/minimalist-gno1 \t[no test files]\n"},
		{args: []string{"test", "../../tests/integ/minimalist-gno2"}, stderrShouldContain: "ok "},
		{args: []string{"test", "../../tests/integ/minimalist-gno3"}, stderrShouldContain: "ok "},
		{args: []string{"test", "../../tests/integ/valid1", "--verbose"}, stderrShouldContain: "ok "},
		{args: []string{"test", "../../tests/integ/valid2"}, stderrShouldContain: "ok "},
		{args: []string{"test", "../../tests/integ/valid2", "--verbose"}, stderrShouldContain: "ok "},

		// TODO: when 'gnodev test' will by default imply running precompile, we should use the following tests.
		//{args: []string{"test", "../../tests/integ/empty-gno1", "--no-precompile"}, stderrShouldBe: "?       ./../../tests/integ/empty-gno1 \t[no test files]\n"},
		//{args: []string{"test", "../../tests/integ/empty-gno1"}, errShouldBe: "FAIL: 1 build errors, 0 test errors", stderrShouldContain: "../../tests/integ/empty-gno1/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'"},
		//{args: []string{"test", "../../tests/integ/empty-gno2", "--no-precompile"}, recoverShouldBe: "empty.gno:1:1: expected 'package', found 'EOF'"}, // FIXME: better error handling + rename dontcare.gno with actual test file
		//{args: []string{"test", "../../tests/integ/empty-gno2"}, errShouldContain: "FAIL: 1 build errors, 0 test errors", stderrShouldContain: "../../tests/integ/empty-gno2/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'"},
		//{args: []string{"test", "../../tests/integ/empty-gno3", "--no-precompile"}, recoverShouldBe: "../../tests/integ/empty-gno3/empty_filetest.gno:1:1: expected 'package', found 'EOF'"}, // FIXME: better error handling
		//{args: []string{"test", "../../tests/integ/empty-gno3"}, errShouldContain: "FAIL: 1 build errors, 0 test errors", stderrShouldContain: "../../tests/integ/empty-gno3/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'"},
		//{args: []string{"test", "../../tests/integ/failing1", "--verbose", "--no-precompile"}, errShouldBe: "FAIL: 0 build errors, 1 test errors", stderrShouldContain: "FAIL: TestAlwaysFailing"},
		//{args: []string{"test", "../../tests/integ/failing1", "--verbose"}, errShouldBe: "FAIL: 0 build errors, 1 test errors", stderrShouldContain: "FAIL: TestAlwaysFailing"},
		//{args: []string{"test", "../../tests/integ/failing2", "--verbose", "--no-precompile"}, recoverShouldBe: "fail on ../../tests/integ/failing2/failing_filetest.gno: got unexpected error: beep boop", stderrShouldContain: "== RUN   file/failing_filetest.gno"},
		//{args: []string{"test", "../../tests/integ/failing2", "--verbose"}, stderrShouldBe: "=== PREC  ./../../tests/integ/failing2\n=== BUILD ./../../tests/integ/failing2\n=== RUN   file/failing_filetest.gno\n", recoverShouldBe: "fail on ../../tests/integ/failing2/failing_filetest.gno: got unexpected error: beep boop"},
		{args: []string{"test", "../../tests/integ/empty-gno1"}, stderrShouldBe: "?       ./../../tests/integ/empty-gno1 \t[no test files]\n"},
		{args: []string{"test", "../../tests/integ/empty-gno1", "--precompile"}, errShouldBe: "FAIL: 1 build errors, 0 test errors", stderrShouldContain: "../../tests/integ/empty-gno1/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'"},
		{args: []string{"test", "../../tests/integ/empty-gno2"}, recoverShouldBe: "empty.gno:1:1: expected 'package', found 'EOF'"}, // FIXME: better error handling + rename dontcare.gno with actual test file
		{args: []string{"test", "../../tests/integ/empty-gno2", "--precompile"}, errShouldContain: "FAIL: 1 build errors, 0 test errors", stderrShouldContain: "../../tests/integ/empty-gno2/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'"},
		{args: []string{"test", "../../tests/integ/empty-gno3"}, recoverShouldBe: "../../tests/integ/empty-gno3/empty_filetest.gno:1:1: expected 'package', found 'EOF'"}, // FIXME: better error handling
		{args: []string{"test", "../../tests/integ/empty-gno3", "--precompile"}, errShouldContain: "FAIL: 1 build errors, 0 test errors", stderrShouldContain: "../../tests/integ/empty-gno3/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'"},
		{args: []string{"test", "../../tests/integ/failing1", "--verbose"}, errShouldBe: "FAIL: 0 build errors, 1 test errors", stderrShouldContain: "FAIL: TestAlwaysFailing"},
		{args: []string{"test", "../../tests/integ/failing1", "--verbose", "--precompile"}, errShouldBe: "FAIL: 0 build errors, 1 test errors", stderrShouldContain: "FAIL: TestAlwaysFailing"},
		{args: []string{"test", "../../tests/integ/failing2", "--verbose"}, recoverShouldBe: "fail on ../../tests/integ/failing2/failing_filetest.gno: got unexpected error: beep boop", stderrShouldContain: "== RUN   file/failing_filetest.gno"},
		{args: []string{"test", "../../tests/integ/failing2", "--verbose", "--precompile"}, stderrShouldBe: "=== PREC  ./../../tests/integ/failing2\n=== BUILD ./../../tests/integ/failing2\n=== RUN   file/failing_filetest.gno\n", recoverShouldBe: "fail on ../../tests/integ/failing2/failing_filetest.gno: got unexpected error: beep boop"},

		// test opts
		{args: []string{"test", "../../examples/gno.land/p/demo/ufmt"}, stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose"}, stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", ".*"}, stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", "NoExists"}, stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", ".*/hello"}, stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", ".*/hi"}, stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", ".*/NoExists"}, stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", ".*/hello/NoExists"}, stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", "Sprintf/"}, stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", "Sprintf/.*"}, stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", "Sprintf/hello"}, stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--timeout", "100000000000" /* 100s */}, stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		// {args: []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--timeout", "10000" /* 10µs */}, recoverShouldContain: "test timed out after"}, // FIXME: should be testable
	}

	for _, test := range tc {
		errShouldBeEmpty := test.errShouldContain == "" && test.errShouldBe == ""
		stdoutShouldBeEmpty := test.stdoutShouldContain == "" && test.stdoutShouldBe == ""
		stderrShouldBeEmpty := test.stderrShouldContain == "" && test.stderrShouldBe == ""
		recoverShouldBeEmpty := test.recoverShouldContain == "" && test.recoverShouldBe == ""

		testName := strings.Join(test.args, " ")
		testName = strings.ReplaceAll(testName, "/", "~")
		t.Run(testName, func(t *testing.T) {
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

				if stdoutShouldBeEmpty {
					require.Empty(t, mockOut.String(), "stdout should be empty")
				} else {
					t.Log("stdout", mockOut.String())
					if test.stdoutShouldContain != "" {
						require.Contains(t, mockOut.String(), test.stdoutShouldContain, "stdout should contain")
					}
					if test.stdoutShouldBe != "" {
						require.Equal(t, mockOut.String(), test.stdoutShouldBe, "stdout should be")
					}
				}

				if stderrShouldBeEmpty {
					require.Empty(t, mockErr.String(), "stderr should be empty")
				} else {
					t.Log("stderr", mockErr.String())
					if test.stderrShouldContain != "" {
						require.Contains(t, mockErr.String(), test.stderrShouldContain, "stderr should contain")
					}
					if test.stderrShouldBe != "" {
						require.Equal(t, mockErr.String(), test.stderrShouldBe, "stderr should be")
					}
				}
			}

			exec := "gnodev"
			defer func() {
				if r := recover(); r != nil {
					output := fmt.Sprintf("%v", r)
					t.Log("recover", output)
					require.False(t, recoverShouldBeEmpty, "should panic")
					require.True(t, errShouldBeEmpty, "should not return an error")
					if test.recoverShouldContain != "" {
						require.Contains(t, output, test.recoverShouldContain, "recover should contain")
					}
					if test.recoverShouldBe != "" {
						require.Equal(t, output, test.recoverShouldBe, "recover should be")
					}
					checkOutputs(t)
				} else {
					require.True(t, recoverShouldBeEmpty, "should not panic")
				}
			}()
			err := runMain(cmd, exec, test.args)

			if errShouldBeEmpty {
				require.Nil(t, err, "err should be nil")
			} else {
				t.Log("err", err.Error())
				require.NotNil(t, err, "err shouldn't be nil")
				if test.errShouldContain != "" {
					require.Contains(t, err.Error(), test.errShouldContain, "err should contain")
				}
				if test.errShouldBe != "" {
					require.Equal(t, err.Error(), test.errShouldBe, "err should be")
				}
			}

			checkOutputs(t)
		})
	}
}
