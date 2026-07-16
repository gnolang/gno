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
			args:             []string{"run", "../../tests/integ/package_mismatched"},
			errShouldContain: "found mismatched packages",
		},
		{
			args:             []string{"run", "../../tests/integ/run_main/main.gno", "../../tests/integ/run_namedpkg"},
			errShouldContain: "found mismatched packages",
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
		{
			args:             []string{"run", "../../tests/integ/undefined_variable/undefined_variables_test.gno"},
			errShouldContain: "gno run: cannot run test files (undefined_variables_test.gno)",
		},
		{
			args:             []string{"run", "../../tests/integ/package_testonly"},
			errShouldContain: "no non-test Gno files in ../../tests/integ/package_testonly",
		},
		{
			// realm package path derived from gnomod.toml; main(cur realm) works.
			args:                []string{"run", "../../tests/integ/run_realm"},
			stdoutShouldContain: "realm main called, counter: 1",
		},
		{
			// -expr calling a crossing function; `cur` is injected.
			args:                []string{"run", "-expr", "Increment()", "../../tests/integ/run_realm"},
			stdoutShouldContain: "counter: 1",
		},
		{
			// realm package path derived from the "// PKGPATH:" directive.
			args:                []string{"run", "../../tests/files/zrealm_std4.gno"},
			stdoutShouldContain: "test2 gno.land/r/tests/vm",
		},
		{
			// realm package path set explicitly with the -pkgpath flag.
			args:                []string{"run", "-pkgpath", "gno.land/r/test/hello", "../../tests/integ/run_main/main.gno"},
			stdoutShouldContain: "hello world!",
		},
		{
			// running local files of a package existing in the store
			// ("package fork").
			args:                []string{"run", "-expr", `println(Render(""))`, "../../../examples/gno.land/r/demo/counter"},
			stdoutShouldContain: "0",
		},
		// TODO: args
		// TODO: nativeLibs VS stdlibs
		// TODO: with gas meter
		// TODO: verbose
		// TODO: logging
	}
	testMainCaseRun(t, tc)
}
