package main

import (
	"errors"
	"go/token"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/require"
)

func TestLintApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"tool", "lint"},
			errShouldBe: "flag: help requested",
		},
		{
			args:                []string{"tool", "lint", "../../tests/integ/run_main/"},
			stderrShouldContain: "./../../tests/integ/run_main: gno.mod file not found in current or any parent directory (code=1)",
			errShouldBe:         "exit code: 1",
		},
		{
			args:                []string{"tool", "lint", "../../tests/integ/undefined_variable_test/undefined_variables_test.gno"},
			stderrShouldContain: "undefined_variables_test.gno:6:28: name toto not declared (code=2)",
			errShouldBe:         "exit code: 1",
		},
		{
			args:                []string{"tool", "lint", "../../tests/integ/package_not_declared/main.gno"},
			stderrShouldContain: "main.gno:4:2: name fmt not declared (code=2)",
			errShouldBe:         "exit code: 1",
		},
		{
			args:                []string{"tool", "lint", "../../tests/integ/several-lint-errors/main.gno"},
			stderrShouldContain: "../../tests/integ/several-lint-errors/main.gno:5:5: expected ';', found example (code=3)\n../../tests/integ/several-lint-errors/main.gno:6",
			errShouldBe:         "exit code: 1",
		},
		{
			args: []string{"tool", "lint", "../../tests/integ/several-files-multiple-errors/main.gno"},
			stderrShouldContain: func() string {
				lines := []string{
					"../../tests/integ/several-files-multiple-errors/file2.gno:3:5: expected 'IDENT', found '{' (code=3)",
					"../../tests/integ/several-files-multiple-errors/file2.gno:5:1: expected type, found '}' (code=3)",
					"../../tests/integ/several-files-multiple-errors/main.gno:5:5: expected ';', found example (code=3)",
					"../../tests/integ/several-files-multiple-errors/main.gno:6:2: expected '}', found 'EOF' (code=3)",
				}
				return strings.Join(lines, "\n") + "\n"
			}(),
			errShouldBe: "exit code: 1",
		},
		{
			args: []string{"tool", "lint", "../../tests/integ/minimalist_gnomod/"},
			// TODO: raise an error because there is a gno.mod, but no .gno files
		},
		{
			args: []string{"tool", "lint", "../../tests/integ/invalid_module_name/"},
			// TODO: raise an error because gno.mod is invalid
		},
		{
			args:                []string{"tool", "lint", "../../tests/integ/invalid_gno_file/"},
			stderrShouldContain: "../../tests/integ/invalid_gno_file/invalid.gno:1:1: expected 'package', found packag (code=2)",
			errShouldBe:         "exit code: 1",
		},
		{
			args:                []string{"tool", "lint", "../../tests/integ/typecheck_missing_return/"},
			stderrShouldContain: "../../tests/integ/typecheck_missing_return/main.gno:5:1: missing return (code=4)",
			errShouldBe:         "exit code: 1",
		},
		{
			args: []string{"tool", "lint", "../../tests/integ/init/"},
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

func TestIssueWithLocationError(t *testing.T) {
	tests := []struct {
		name  string
		inErr error
		want  lintIssue
	}{
		{
			"nil error",
			nil,
			lintIssue{
				Code:       lintGnoError,
				Confidence: 1,
			},
		},
		{
			"location in error",
			gnolang.MakeLocationPlusError(
				token.Position{"foo.gno", 0, 1, 1},
				"this panicked",
			),
			lintIssue{
				Code:       lintGnoError,
				Confidence: 1,
				Msg:        "this panicked",
				Location:   "foo.gno:1:1",
			},
		},
		{
			"generic error",
			errors.New("foo.gno:10:18: this one"),
			lintIssue{
				Code:       lintGnoError,
				Confidence: 1,
				Msg:        "this one",
				Location:   "path:10:18",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := issueFromError("path", tt.inErr)
			require.Equal(t, got, tt.want)
		})
	}
}
