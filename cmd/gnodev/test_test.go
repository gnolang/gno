package main

import "testing"

func TestTest(t *testing.T) {
	tc := []testMainCase{
		{
			args:           []string{"test"},
			errShouldBe:    "invalid args",
			stderrShouldBe: "Usage: test [test flags] [packages]\n",
		}, {
			args:                []string{"test", "--help"},
			stdoutShouldContain: "# testOptions options\n-",
		}, {
			args:                []string{"test", "../../examples/gno.land/p/demo/rand"},
			stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/rand \t",
		}, {
			args:             []string{"test", "../../tests/integ/no-such-dir"},
			errShouldContain: "no such file or directory",
		}, {
			args: []string{"test", "../../tests/integ/empty-dir"},
		}, {
			// FIXME: should have an output
			args:           []string{"test", "../../tests/integ/minimalist-gno1"},
			stderrShouldBe: "?       ./../../tests/integ/minimalist-gno1 \t[no test files]\n",
		}, {
			args:                []string{"test", "../../tests/integ/minimalist-gno2"},
			stderrShouldContain: "ok ",
		}, {
			args:                []string{"test", "../../tests/integ/minimalist-gno3"},
			stderrShouldContain: "ok ",
		}, {
			args:                []string{"test", "../../tests/integ/valid1", "--verbose"},
			stderrShouldContain: "ok ",
		}, {
			args:                []string{"test", "../../tests/integ/valid2"},
			stderrShouldContain: "ok ",
		}, {
			args:                []string{"test", "../../tests/integ/valid2", "--verbose"},
			stderrShouldContain: "ok ",
		}, {
			args:           []string{"test", "../../tests/integ/empty-gno1"},
			stderrShouldBe: "?       ./../../tests/integ/empty-gno1 \t[no test files]\n",
		}, {
			args:                []string{"test", "../../tests/integ/empty-gno1", "--precompile"},
			errShouldBe:         "FAIL: 1 build errors, 0 test errors",
			stderrShouldContain: "../../tests/integ/empty-gno1/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'",
		}, {
			args:            []string{"test", "../../tests/integ/empty-gno2"},
			recoverShouldBe: "empty.gno:1:1: expected 'package', found 'EOF'",
		}, {
			// FIXME: better error handling + rename dontcare.gno with actual test file
			args:                []string{"test", "../../tests/integ/empty-gno2", "--precompile"},
			errShouldContain:    "FAIL: 1 build errors, 0 test errors",
			stderrShouldContain: "../../tests/integ/empty-gno2/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'",
		}, {
			args:            []string{"test", "../../tests/integ/empty-gno3"},
			recoverShouldBe: "../../tests/integ/empty-gno3/empty_filetest.gno:1:1: expected 'package', found 'EOF'",
		}, {
			// FIXME: better error handling
			args:                []string{"test", "../../tests/integ/empty-gno3", "--precompile"},
			errShouldContain:    "FAIL: 1 build errors, 0 test errors",
			stderrShouldContain: "../../tests/integ/empty-gno3/empty.gno: parse: tmp.gno:1:1: expected 'package', found 'EOF'",
		}, {
			args:                []string{"test", "../../tests/integ/failing1", "--verbose"},
			errShouldBe:         "FAIL: 0 build errors, 1 test errors",
			stderrShouldContain: "FAIL: TestAlwaysFailing",
		}, {
			args:                []string{"test", "../../tests/integ/failing1", "--verbose", "--precompile"},
			errShouldBe:         "FAIL: 0 build errors, 1 test errors",
			stderrShouldContain: "FAIL: TestAlwaysFailing",
		}, {
			args:                []string{"test", "../../tests/integ/failing2", "--verbose"},
			recoverShouldBe:     "fail on ../../tests/integ/failing2/failing_filetest.gno: got unexpected error: beep boop",
			stderrShouldContain: "== RUN   file/failing_filetest.gno",
		}, {
			args:            []string{"test", "../../tests/integ/failing2", "--verbose", "--precompile"},
			stderrShouldBe:  "=== PREC  ./../../tests/integ/failing2\n=== BUILD ./../../tests/integ/failing2\n=== RUN   file/failing_filetest.gno\n",
			recoverShouldBe: "fail on ../../tests/integ/failing2/failing_filetest.gno: got unexpected error: beep boop",
		}, {
			args:                []string{"test", "../../examples/gno.land/p/demo/ufmt"},
			stdoutShouldContain: "RUN   TestSprintf",
			stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt",
		}, {
			args:                []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose"},
			stdoutShouldContain: "RUN   TestSprintf",
			stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt",
		}, {
			args:                []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", ".*"},
			stdoutShouldContain: "RUN   TestSprintf",
			stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt",
		}, {
			args:                []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", "NoExists"},
			stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt",
		}, {
			args:                []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", ".*/hello"},
			stdoutShouldContain: "RUN   TestSprintf",
			stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt",
		}, {
			args:                []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", ".*/hi"},
			stdoutShouldContain: "RUN   TestSprintf",
			stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt",
		}, {
			args:                []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", ".*/NoExists"},
			stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt",
		}, {
			args:                []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", ".*/hello/NoExists"},
			stdoutShouldContain: "RUN   TestSprintf",
			stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt",
		}, {
			args:                []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", "Sprintf/"},
			stdoutShouldContain: "RUN   TestSprintf",
			stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt",
		}, {
			args:                []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", "Sprintf/.*"},
			stdoutShouldContain: "RUN   TestSprintf",
			stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt",
		}, {
			args:                []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--run", "Sprintf/hello"},
			stdoutShouldContain: "RUN   TestSprintf",
			stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt",
		}, {
			args:                []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--timeout", "100000000000" /* 100s */},
			stdoutShouldContain: "RUN   TestSprintf",
			stderrShouldContain: "ok      ./../../examples/gno.land/p/demo/ufmt",
		},

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
		// {args: []string{"test", "../../examples/gno.land/p/demo/ufmt", "--verbose", "--timeout", "10000" /* 10µs */}, recoverShouldContain: "test timed out after"}, // FIXME: should be testable
	}
	testMainCaseRun(t, tc)
}
