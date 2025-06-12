package main

import (
	"strings"
	"testing"
)

func TestLintApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"lint"},
			errShouldBe: "flag: help requested",
		},
		{
			args:                []string{"lint", "../../tests/integ/run_main/", "-auto-gnomod=false"},
			stderrShouldContain: "../../tests/integ/run_main/gno.mod: could not read gno.mod file: stat ../../tests/integ/run_main/gno.mod: no such file or directory (code=gnoGnoModError)",
			errShouldBe:         "exit code: 1",
		},
		{
			args:                []string{"lint", "../../tests/integ/undefined_variable_test/undefined_variables_test.gno"},
			stderrShouldContain: "../../tests/integ/undefined_variable_test/undefined_variables_test.gno:6:28: undefined: toto (code=gnoTypeCheckError)",
			errShouldBe:         "exit code: 1",
		},
		{
			args:                []string{"lint", "../../tests/integ/package_not_declared/main.gno"},
			stderrShouldContain: "../../tests/integ/package_not_declared/main.gno:4:2: undefined: fmt (code=gnoTypeCheckError)",
			errShouldBe:         "exit code: 1",
		},
		{
			args:                []string{"lint", "../../tests/integ/several-lint-errors/main.gno"},
			stderrShouldContain: "../../tests/integ/several-lint-errors/main.gno:5:5: expected ';', found example (code=gnoParserError)\n../../tests/integ/several-lint-errors/main.gno:6",
			errShouldBe:         "exit code: 1",
		},
		{
			args: []string{"lint", "../../tests/integ/several-files-multiple-errors/main.gno"},
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
			args: []string{"lint", "../../tests/integ/minimalist_gnomod/"},
			// TODO: raise an error because there is a gno.mod, but no .gno files
		},
		{
			args: []string{"lint", "../../tests/integ/invalid_module_name/"},
			// TODO: raise an error because gno.mod is invalid
		},
		{
			args:                []string{"lint", "../../tests/integ/invalid_gno_file/"},
			stderrShouldContain: "../../tests/integ/invalid_gno_file/invalid.gno:1:1: expected 'package', found packag (code=gnoParserError)",
			errShouldBe:         "exit code: 1",
		},
		{
			args:                []string{"lint", "../../tests/integ/typecheck_missing_return/"},
			stderrShouldContain: "../../tests/integ/typecheck_missing_return/main.gno:5:1: missing return (code=gnoTypeCheckError)",
			errShouldBe:         "exit code: 1",
		},
		{
			args: []string{"lint", "../../tests/integ/init/"},
			// stderr / stdout should be empty; the init function and statements
			// should not be executed
		},

		// TODO: 'gno mod' is valid?
		// TODO: are dependencies valid?
		// TODO: is gno source using unsafe/discouraged features?
		// TODO: check for imports of native libs from non _test.gno files
	}
	testMainCaseRun(t, tc)
}
