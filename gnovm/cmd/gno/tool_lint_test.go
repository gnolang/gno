package main

import (
	"strings"
	"testing"
)

func TestLintApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"tool", "lint"},
			errShouldBe: "flag: help requested",
		}, {
			args:                []string{"tool", "lint", "../../tests/integ/run_main/"},
			stderrShouldContain: "../../tests/integ/run_main: gno.mod file not found in current or any parent directory (code=1)",
			errShouldBe:         "exit code: 1",
		}, {
			args:                []string{"tool", "lint", "../../tests/integ/undefined_variable_test/undefined_variables_test.gno"},
			stderrShouldContain: "undefined_variables_test.gno:6:28: name toto not declared (code=2)",
			errShouldBe:         "exit code: 1",
		}, {
			args:                []string{"tool", "lint", "../../tests/integ/package_not_declared/main.gno"},
			stderrShouldContain: "../../tests/integ/package_not_declared/main.gno:4:2: undefined: fmt (code=4)\n",
			errShouldBe:         "exit code: 1",
		}, {
			args:           []string{"tool", "lint", "../../tests/integ/several-lint-errors/main.gno"},
			stderrShouldBe: "../../tests/integ/several-lint-errors/main.gno:5:5: expected ';', found example (code=3)\n../../tests/integ/several-lint-errors/main.gno:6:2: expected '}', found 'EOF' (code=3)\n",
			errShouldBe:    "exit code: 1",
		}, {
			args: []string{"tool", "lint", "../../tests/integ/several-files-multiple-errors/main.gno"},
			stderrShouldContain: func() string {
				lines := []string{
					"../../tests/integ/several-files-multiple-errors/file2.gno:3:5: expected 'IDENT', found '{' (code=2)",
					"../../tests/integ/several-files-multiple-errors/file2.gno:5:1: expected type, found '}' (code=2)",
					"../../tests/integ/several-files-multiple-errors/main.gno:5:5: expected ';', found example (code=2)",
					"../../tests/integ/several-files-multiple-errors/main.gno:6:2: expected '}', found 'EOF' (code=2)",
				}
				return strings.Join(lines, "\n") + "\n"
			}(),
			errShouldBe: "exit code: 1",
		}, {
			args: []string{"tool", "lint", "../../tests/integ/minimalist_gnomod/"},
			// TODO: raise an error because there is a gno.mod, but no .gno files
		}, {
			args:           []string{"tool", "lint", "../../tests/integ/invalid_module_name/"},
			stderrShouldBe: "../../tests/integ/invalid_module_name/gno.mod:1: usage: module module/path (code=5)\n",
			errShouldBe:    "exit code: 1",
		}, {
			args:           []string{"tool", "lint", "../../tests/integ/invalid_gno_file/"},
			stderrShouldBe: "../../tests/integ/invalid_gno_file/invalid.gno:1:1: expected 'package', found packag (code=5)\n",
			errShouldBe:    "exit code: 1",
		}, {
			args:           []string{"tool", "lint", "../../tests/integ/typecheck_missing_return/"},
			stderrShouldBe: "../../tests/integ/typecheck_missing_return/main.gno:5:1: missing return (code=4)\n",
			errShouldBe:    "exit code: 1",
		},

		// TODO: 'gno mod' is valid?
		// TODO: are dependencies valid?
		// TODO: is gno source using unsafe/discouraged features?
		// TODO: check for imports of native libs from non _test.gno files
	}
	testMainCaseRun(t, tc)
}
