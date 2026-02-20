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
			stderrShouldBe:       "undefined_variables_test.gno:6:28: undefined: toto (code=gnoTypeCheckError)\n",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/package_not_declared",
			simulateExternalRepo: true,
			stderrShouldBe:       "main.gno:4:2: undefined: fmt (code=gnoTypeCheckError)\n",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/several-lint-errors",
			simulateExternalRepo: true,
			stderrShouldBe:       "main.gno:5:5: expected ';', found example (code=gnoParserError)\nmain.gno:6:2: expected '}', found 'EOF' (code=gnoParserError)\n",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/several-files-multiple-errors",
			simulateExternalRepo: true,
			stderrShouldBe: func() string {
				lines := []string{
					"file2.gno:3:5: expected 'IDENT', found '{' (code=gnoParserError)",
					"file2.gno:5:1: expected type, found '}' (code=gnoParserError)",
					"main.gno:5:5: expected ';', found example (code=gnoParserError)",
					"main.gno:6:2: expected '}', found 'EOF' (code=gnoParserError)",
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
			stderrShouldContain:  "gnomod.toml: invalid gnomod.toml: 'module' is required (code=gnoGnoModError)",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/invalid_gno_file",
			simulateExternalRepo: true,
			stderrShouldBe:       "invalid.gno:1:1: expected 'package', found packag (code=gnoParserError)\n",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/typecheck_missing_return",
			simulateExternalRepo: true,
			stderrShouldBe:       "main.gno:5:1: missing return (code=gnoTypeCheckError)\n",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/init",
			simulateExternalRepo: true,
			// stderr / stdout should be empty; the init function and statements
			// should not be executed
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/package_name_mismatch",
			simulateExternalRepo: true,
			stderrShouldContain:  `package name "hello" does not match path element "goodbye" (code=gnoPackageNameMismatch)`,
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/render_invalid1",
			simulateExternalRepo: true,
			stderrShouldBe:       "gno.land/r/test/render_invalid1/main.gno:5: invalid signature for the realm's Render function; must be of the form: func Render(string) string (code=gnoLintError)\n",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/render_invalid2",
			simulateExternalRepo: true,
			stderrShouldBe:       "gno.land/r/test/render_invalid2/main.gno:5: invalid signature for the realm's Render function; must be of the form: func Render(string) string (code=gnoLintError)\n",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/render_invalid3",
			simulateExternalRepo: true,
			stderrShouldBe:       "gno.land/r/test/render_invalid3/main.gno:5: invalid signature for the realm's Render function; must be of the form: func Render(string) string (code=gnoLintError)\n",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/render_invalid4",
			simulateExternalRepo: true,
			stderrShouldBe:       "gno.land/r/test/render_invalid4/main.gno:5: invalid signature for the realm's Render function; must be of the form: func Render(string) string (code=gnoLintError)\n",
			errShouldBe:          "exit code: 1",
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/render_valid1",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/render_valid2",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/render_valid3",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"lint", "."},
			testDir:              "../../tests/integ/render_valid4",
			simulateExternalRepo: true,
		},

		// TODO: 'gno mod' is valid?
		// TODO: are dependencies valid?
		// TODO: is gno source using unsafe/discouraged features?
		// TODO: check for imports of native libs from non _test.gno files
	}
	testMainCaseRun(t, tc)
}
