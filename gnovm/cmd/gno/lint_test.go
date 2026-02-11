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
			args:                 []string{"lint", ".", "-auto-gnomod=false"},
			testDir:              "../../tests/integ/run_main",
			simulateExternalRepo: true,
			errShouldBe:          "gnowork.toml file not found in current or any parent directory and gnomod.toml doesn't exists in current directory",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/undefined_variable",
			simulateExternalRepo: true,
			stderrShouldBe:       "undefined_variables_test.gno:6:28: error: undefined: toto (gnoTypeCheckError)\n\nFound 1 issue(s): 1 error(s), 0 warning(s), 0 info\n",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/package_not_declared",
			simulateExternalRepo: true,
			stderrShouldBe:       "main.gno:4:2: error: undefined: fmt (gnoTypeCheckError)\n\nFound 1 issue(s): 1 error(s), 0 warning(s), 0 info\n",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/several-lint-errors",
			simulateExternalRepo: true,
			stderrShouldBe:       "main.gno:5:5: error: expected ';', found example (gnoParserError)\nmain.gno:6:2: error: expected '}', found 'EOF' (gnoParserError)\n\nFound 2 issue(s): 2 error(s), 0 warning(s), 0 info\n",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/several-files-multiple-errors",
			simulateExternalRepo: true,
			stderrShouldBe: func() string {
				lines := []string{
					"file2.gno:3:5: error: expected 'IDENT', found '{' (gnoParserError)",
					"file2.gno:5:1: error: expected type, found '}' (gnoParserError)",
					"main.gno:5:5: error: expected ';', found example (gnoParserError)",
					"main.gno:6:2: error: expected '}', found 'EOF' (gnoParserError)",
					"",
					"Found 4 issue(s): 4 error(s), 0 warning(s), 0 info",
				}
				return strings.Join(lines, "\n") + "\n"
			}(),
			errShouldBe: "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/minimalist_gnomod",
			simulateExternalRepo: true,
			// TODO: raise an error because there is a gno.mod, but no .gno files
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/invalid_module_name",
			simulateExternalRepo: true,
			stderrShouldContain:  "gnomod.toml:0:0: error: invalid gnomod.toml: 'module' is required (gnoGnoModError)",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/invalid_gno_file",
			simulateExternalRepo: true,
			stderrShouldBe:       "invalid.gno:1:1: error: expected 'package', found packag (gnoParserError)\n\nFound 1 issue(s): 1 error(s), 0 warning(s), 0 info\n",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/typecheck_missing_return",
			simulateExternalRepo: true,
			stderrShouldBe:       "main.gno:5:1: error: missing return (gnoTypeCheckError)\n\nFound 1 issue(s): 1 error(s), 0 warning(s), 0 info\n",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/init",
			simulateExternalRepo: true,
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
