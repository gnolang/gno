package main

import "testing"

func TestLintApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"lint"},
			errShouldBe: "flag: help requested",
		},
		{
			args:                []string{"lint", "../../tests/integ/lint-main/main.gno"},
			stdoutShouldContain: "hello world!",
		},
		{
			args:                 []string{"lint", "../../tests/integ/lint-main/"},
			recoverShouldContain: "read ../../tests/integ/lint-main/: is a directory", // FIXME: should work
		},
		{
			args:                 []string{"lint", "../../tests/integ/does-not-exist"},
			recoverShouldContain: "open ../../tests/integ/does-not-exist: no such file or directory",
		},
		{
			args:                 []string{"lint", "../../tests/integ/lint-namedpkg/main.gno"},
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
	testMainCaseLint(t, tc)
}
