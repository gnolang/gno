package main

import "testing"

func TestRunApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"run"},
			errShouldBe: "flag: help requested",
		},
		{
			args:                []string{"run", "../../tests/integ/run-main/main.gno"},
			stdoutShouldContain: "hello world!",
		},
		{
			args:                 []string{"run", "../../tests/integ/run-main/"},
			recoverShouldContain: "read ../../tests/integ/run-main/: is a directory", // FIXME: should work
		},
		{
			args:                 []string{"run", "../../tests/integ/does-not-exist"},
			recoverShouldContain: "open ../../tests/integ/does-not-exist: no such file or directory",
		},
		{
			args:                 []string{"run", "../../tests/integ/run-namedpkg/main.gno"},
			recoverShouldContain: "expected package name [main] but got [namedpkg]", // FIXME: should work
		},
		// TODO: multiple files
		// TODO: a test file
		// TODO: a file without main
		// TODO: args
		// TODO: nativeLibs VS stdlibs
		// TODO: with gas meter
		// TODO: verbose
		// TODO: logging
	}
	testMainCaseRun(t, tc)
}
