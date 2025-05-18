package main

import (
	"strings"
	"testing"
)

func TestRunApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"run"},
			errShouldBe: "flag: help requested",
		},
		{
			args:                []string{"run", "../../tests/integ/run_main/main.gno"},
			stdoutShouldContain: "hello world!",
		},
		{
			args:                []string{"run", "../../tests/integ/run_main/"},
			stdoutShouldContain: "hello world!",
		},
		{
			args:             []string{"run", "../../tests/integ/does_not_exist"},
			errShouldContain: "no such file or directory",
		},
		{
			args:                []string{"run", "../../tests/integ/run_namedpkg/main.gno"},
			stdoutShouldContain: "hello, other world!",
		},
		{
			args:             []string{"run", "../../tests/integ/run_package"},
			errShouldContain: "name main not declared",
		},
		{
			args:                []string{"run", "-expr", "Hello()", "../../tests/integ/run_package"},
			stdoutShouldContain: "called Hello",
		},
		{
			args:                []string{"run", "-expr", "World()", "../../tests/integ/run_package"},
			stdoutShouldContain: "called World",
		},
		{
			args:                []string{"run", "-expr", "otherFile()", "../../tests/integ/run_package"},
			stdoutShouldContain: "hello from package2.gno",
		},
		{
			args: []string{
				"run", "-expr", "otherFile()",
				"../../tests/integ/run_package/package.gno",
			},
			errShouldContain: "name otherFile not declared",
		},
		{
			args: []string{
				"run", "-expr", "otherFile()",
				"../../tests/integ/run_package/package.gno",
				"../../tests/integ/run_package/package2.gno",
			},
			stdoutShouldContain: "hello from package2.gno",
		},
		{
			args:                []string{"run", "-expr", "WithArg(1)", "../../tests/integ/run_package"},
			stdoutShouldContain: "one",
		},
		{
			args:                []string{"run", "-expr", "WithArg(-255)", "../../tests/integ/run_package"},
			stdoutShouldContain: "out of range!",
		},
		{
			args:                 []string{"run", "../../tests/integ/undefined_variable_test/undefined_variables_test.gno"},
			recoverShouldContain: "--- preprocess stack ---", // should contain preprocess debug stack trace
		},
		{
			args:                []string{"run", "-debug", "../../tests/integ/debugger/sample.gno"},
			stdoutShouldContain: "Welcome to the Gnovm debugger",
		},
		{
			args:             []string{"run", "-debug-addr", "invalidhost:17538", "../../tests/integ/debugger/sample.gno"},
			errShouldContain: "listen tcp",
		},
		{
			args:                 []string{"run", "../../tests/integ/invalid_assign/main.gno"},
			recoverShouldContain: "cannot use bool as main.C without explicit conversion",
		},
		{
			args:                []string{"run", "-expr", "Context()", "../../tests/integ/context/context.gno"},
			stdoutShouldContain: "Context worked",
		},
		{
			args: []string{"run", "../../tests/integ/several-files-multiple-errors/"},
			stderrShouldContain: func() string {
				lines := []string{
					"../../tests/integ/several-files-multiple-errors/file2.gno:3:5: expected 'IDENT', found '{' (code=gnoParserError)",
					"../../tests/integ/several-files-multiple-errors/file2.gno:5:1: expected type, found '}' (code=gnoParserError)",
					"../../tests/integ/several-files-multiple-errors/main.gno:5:5: expected ';', found example (code=gnoParserError)",
					"../../tests/integ/several-files-multiple-errors/main.gno:6:2: expected '}', found 'EOF' (code=gnoParserError)",
				}
				return strings.Join(lines, "\n") + "\n"
			}(),
			errShouldBe: "exit code: 1",
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
