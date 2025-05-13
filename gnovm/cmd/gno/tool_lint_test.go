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
			stderrShouldContain: "../../tests/integ/run_main/gno.mod: could not read gno.mod file: stat ../../tests/integ/run_main/gno.mod: no such file or directory (code=lintGnoModError)",
			errShouldBe:         "exit code: 1",
		},
		{
			args:                []string{"lint", "../../tests/integ/undefined_variable_test/undefined_variables_test.gno"},
			stderrShouldContain: "undefined_variables_test.gno:6:28: name toto not declared (code=lintPreprocessError)",
			errShouldBe:         "exit code: 1",
		},
		{
			args:                []string{"lint", "../../tests/integ/package_not_declared/main.gno"},
			stderrShouldContain: "main.gno:4:2: name fmt not declared (code=lintPreprocessError)",
			errShouldBe:         "exit code: 1",
		},
		{
			args:                []string{"lint", "../../tests/integ/several-lint-errors/main.gno"},
			stderrShouldContain: "../../tests/integ/several-lint-errors/main.gno:5:5: expected ';', found example (code=lintParserError)\n../../tests/integ/several-lint-errors/main.gno:6",
			errShouldBe:         "exit code: 1",
		},
		{
			args: []string{"lint", "../../tests/integ/several-files-multiple-errors/main.gno"},
			stderrShouldContain: func() string {
				lines := []string{
					"../../tests/integ/several-files-multiple-errors/file2.gno:3:5: expected 'IDENT', found '{' (code=lintParserError)",
					"../../tests/integ/several-files-multiple-errors/file2.gno:5:1: expected type, found '}' (code=lintParserError)",
					"../../tests/integ/several-files-multiple-errors/main.gno:5:5: expected ';', found example (code=lintParserError)",
					"../../tests/integ/several-files-multiple-errors/main.gno:6:2: expected '}', found 'EOF' (code=lintParserError)",
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
			stderrShouldContain: "../../tests/integ/invalid_gno_file/invalid.gno:1:1: expected 'package', found packag (code=lintReadError)",
			errShouldBe:         "exit code: 1",
		},
		{
			args:                []string{"lint", "../../tests/integ/typecheck_missing_return/"},
			stderrShouldContain: "../../tests/integ/typecheck_missing_return/main.gno:5:1: missing return (code=lintTypeCheckError)",
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
