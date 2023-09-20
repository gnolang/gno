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
			args:                []string{"run", "../../tests/integ/run-main/"},
			stdoutShouldContain: "hello world!",
		},
		{
			args:             []string{"run", "../../tests/integ/does-not-exist"},
			errShouldContain: "no such file or directory",
		},
		{
			args:                []string{"run", "../../tests/integ/run-namedpkg/main.gno"},
			stdoutShouldContain: "hello, other world!",
		},
		{
			args:                 []string{"run", "../../tests/integ/run-package"},
			recoverShouldContain: "name main not declared",
		},
		{
			args:                []string{"run", "-expr", "Hello()", "../../tests/integ/run-package"},
			stdoutShouldContain: "called Hello",
		},
		{
			args:                []string{"run", "-expr", "World()", "../../tests/integ/run-package"},
			stdoutShouldContain: "called World",
		},
		{
			args:                []string{"run", "-expr", "otherFile()", "../../tests/integ/run-package"},
			stdoutShouldContain: "hello from package2.gno",
		},
		{
			args: []string{
				"run", "-expr", "otherFile()",
				"../../tests/integ/run-package/package.gno",
			},
			recoverShouldContain: "name otherFile not declared",
		},
		{
			args: []string{
				"run", "-expr", "otherFile()",
				"../../tests/integ/run-package/package.gno",
				"../../tests/integ/run-package/package2.gno",
			},
			stdoutShouldContain: "hello from package2.gno",
		},
		{
			args:                []string{"run", "-expr", "WithArg(1)", "../../tests/integ/run-package"},
			stdoutShouldContain: "one",
		},
		{
			args:                []string{"run", "-expr", "WithArg(-255)", "../../tests/integ/run-package"},
			stdoutShouldContain: "out of range!",
		},
		// TODO: a test file
		// TODO: args
		// TODO: nativeLibs VS stdlibs
		// TODO: with gas meter
		// TODO: verbose
		// TODO: logging
	}
	testMainCaseRun(t, tc)
}
