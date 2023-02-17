package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain_Run(t *testing.T) {
	t.SkipNow()

	tc := []struct {
		args []string

		// for the following FooContain+FooBe expected couples, if both are empty,
		// then the test suite will require that the "got" is not empty.
		errShouldContain    string
		errShouldBe         string
		stdoutShouldBe      string
		stdoutShouldContain string
	}{
		// custom
		{args: []string{"test", "--verbose", "../../examples/gno.land/p/demo/rand"}, stdoutShouldContain: "ok      ./../../examples/gno.land/p/demo/rand \t"},
		{args: []string{"test", "../../tests/integ/no-such-dir"}, errShouldContain: "no such file or directory"},
		{args: []string{"test", "../../tests/integ/empty-dir"}}, // FIXME: should have an output
		{args: []string{"test", "../../tests/integ/minimalist-gno1"}, stdoutShouldBe: "?       ./../../tests/integ/minimalist-gno1 \t[no test files]\n"},
		{args: []string{"test", "../../tests/integ/minimalist-gno2"}, stdoutShouldContain: "ok "},
		{args: []string{"test", "--verbose", "../../tests/integ/minimalist-gno3"}, stdoutShouldContain: "ok "},
		{args: []string{"test", "--verbose", "../../tests/integ/valid1"}, stdoutShouldContain: "ok "},
		{args: []string{"test", "--verbose", "../../tests/integ/valid2"}, stdoutShouldContain: "ok "},
		{args: []string{"test", "--verbose", "../../tests/integ/valid2"}, stdoutShouldContain: "ok "},

		// TODO: when 'gnodev test' will by default imply running precompile, we should use the following tests.
		// {args: []string{"test", "../../tests/integ/empty-gno1", "--no-precompile"}, stderrShouldBe: "?       ./../../tests/integ/empty-gno1 \t[no test files]\n"},
		// {args: []string{"test", "../../tests/integ/empty-gno1"}, errShouldBe: "FAIL: 1 build errors, 0 test errors", stderrShouldContain: "../../tests/integ/empty-gno1/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'"},
		// {args: []string{"test", "../../tests/integ/empty-gno2", "--no-precompile"}, recoverShouldBe: "empty.gno:1:1: expected 'package', found 'EOF'"}, // FIXME: better error handling + rename dontcare.gno with actual test file
		// {args: []string{"test", "../../tests/integ/empty-gno2"}, errShouldContain: "FAIL: 1 build errors, 0 test errors", stderrShouldContain: "../../tests/integ/empty-gno2/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'"},
		// {args: []string{"test", "../../tests/integ/empty-gno3", "--no-precompile"}, recoverShouldBe: "../../tests/integ/empty-gno3/empty_filetest.gno:1:1: expected 'package', found 'EOF'"}, // FIXME: better error handling
		// {args: []string{"test", "../../tests/integ/empty-gno3"}, errShouldContain: "FAIL: 1 build errors, 0 test errors", stderrShouldContain: "../../tests/integ/empty-gno3/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'"},
		// {args: []string{"test", "../../tests/integ/failing1", "--verbose", "--no-precompile"}, errShouldBe: "FAIL: 0 build errors, 1 test errors", stderrShouldContain: "FAIL: TestAlwaysFailing"},
		// {args: []string{"test", "../../tests/integ/failing1", "--verbose"}, errShouldBe: "FAIL: 0 build errors, 1 test errors", stderrShouldContain: "FAIL: TestAlwaysFailing"},
		// {args: []string{"test", "../../tests/integ/failing2", "--verbose", "--no-precompile"}, recoverShouldBe: "fail on ../../tests/integ/failing2/failing_filetest.gno: got unexpected error: beep boop", stderrShouldContain: "== RUN   file/failing_filetest.gno"},
		// {args: []string{"test", "../../tests/integ/failing2", "--verbose"}, stderrShouldBe: "=== PREC  ./../../tests/integ/failing2\n=== BUILD ./../../tests/integ/failing2\n=== RUN   file/failing_filetest.gno\n", recoverShouldBe: "fail on ../../tests/integ/failing2/failing_filetest.gno: got unexpected error: beep boop"},
		{args: []string{"test", "../../tests/integ/empty-gno1"}, stdoutShouldBe: "?       ./../../tests/integ/empty-gno1 \t[no test files]\n"},
		{args: []string{"test", "--precompile", "../../tests/integ/empty-gno1"}, errShouldBe: "FAIL: 1 build errors, 0 test errors", stdoutShouldContain: "../../tests/integ/empty-gno1/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'"},
		{args: []string{"test", "../../tests/integ/empty-gno2"}, errShouldContain: "empty.gno:1:1: expected 'package', found 'EOF'"}, // FIXME: better error handling + rename dontcare.gno with actual test file
		{args: []string{"test", "--precompile", "../../tests/integ/empty-gno2"}, errShouldContain: "FAIL: 1 build errors, 0 test errors", stdoutShouldContain: "../../tests/integ/empty-gno2/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'"},
		{args: []string{"test", "../../tests/integ/empty-gno3"}, errShouldContain: "../../tests/integ/empty-gno3/empty_filetest.gno:1:1: expected 'package', found 'EOF'"}, // FIXME: better error handling
		{args: []string{"test", "--precompile", "../../tests/integ/empty-gno3"}, errShouldContain: "FAIL: 1 build errors, 0 test errors", stdoutShouldContain: "../../tests/integ/empty-gno3/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'"},
		{args: []string{"test", "--verbose", "../../tests/integ/failing1"}, errShouldBe: "FAIL: 0 build errors, 1 test errors", stdoutShouldContain: "FAIL: TestAlwaysFailing"},
		{args: []string{"test", "--verbose", "--precompile", "../../tests/integ/failing1"}, errShouldBe: "FAIL: 0 build errors, 1 test errors", stdoutShouldContain: "FAIL: TestAlwaysFailing"},
		{args: []string{"test", "--verbose", "../../tests/integ/failing2"}, errShouldContain: "fail on ../../tests/integ/failing2/failing_filetest.gno: got unexpected error: beep boop", stdoutShouldContain: "== RUN   file/failing_filetest.gno"},

		// test opts
		{args: []string{"test", "../../examples/gno.land/p/demo/ufmt"}, stdoutShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "--verbose", "../../examples/gno.land/p/demo/ufmt"}, stdoutShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "--verbose", "--run", ".*", "../../examples/gno.land/p/demo/ufmt"}, stdoutShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "--verbose", "--run", "NoExists", "../../examples/gno.land/p/demo/ufmt"}, stdoutShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "--verbose", "--run", ".*/hello", "../../examples/gno.land/p/demo/ufmt"}, stdoutShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "--verbose", "--run", ".*/hi", "../../examples/gno.land/p/demo/ufmt"}, stdoutShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "--verbose", "--run", ".*/NoExists", "../../examples/gno.land/p/demo/ufmt"}, stdoutShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "--verbose", "--run", ".*/hello/NoExists", "../../examples/gno.land/p/demo/ufmt"}, stdoutShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "--verbose", "--run", "Sprintf/", "../../examples/gno.land/p/demo/ufmt"}, stdoutShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "--verbose", "--run", "Sprintf/.*", "../../examples/gno.land/p/demo/ufmt"}, stdoutShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "--verbose", "--run", "Sprintf/hello", "../../examples/gno.land/p/demo/ufmt"}, stdoutShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		{args: []string{"test", "--verbose", "--timeout", "100000s", "../../examples/gno.land/p/demo/ufmt"}, stdoutShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt"},
		// {args: []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--timeout", "10000" /* 10µs */}, recoverShouldContain: "test timed out after"}, // FIXME: should be testable
	}

	for _, test := range tc {
		errShouldBeEmpty := test.errShouldContain == "" && test.errShouldBe == ""
		stdoutShouldBeEmpty := test.stdoutShouldContain == "" && test.stdoutShouldBe == ""

		testName := strings.Join(test.args, " ")
		testName = strings.ReplaceAll(testName, "/", "~")
		t.Run(testName, func(t *testing.T) {
			mockOut := bytes.NewBufferString("")
			mockErr := bytes.NewBufferString("")
			// stdout := command.WriteNopCloser(mockOut)
			// stderr := command.WriteNopCloser(mockErr)

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
						require.Equal(t, test.stdoutShouldBe, mockOut.String(), "stdout should be")
					}
				}
			}

			execCmd := exec.Command("gnodev", test.args...)
			// execCmd.Stdout = stdout
			// execCmd.Stderr = stderr

			err := execCmd.Run()

			if errShouldBeEmpty {
				require.Nil(t, err, "err should be nil")
			} else {
				require.NotNil(t, err, "err shouldn't be nil")
				if test.errShouldContain != "" {
					require.Contains(t, mockErr.String(), test.errShouldContain, "err should contain")
				}
				if test.errShouldBe != "" {
					require.Equal(t, test.errShouldBe, mockErr.String(), "err should be")
				}
			}

			checkOutputs(t)
		})
	}
}
