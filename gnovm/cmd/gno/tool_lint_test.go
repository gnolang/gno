package main

import (
	"bytes"
	"go/types"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/commands"
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

func TestLintRenderSignature(t *testing.T) {
	type test struct {
		desc  string
		input string
		err   bool
	}

	const errMsg = "gno.land/test: The 'Render' function signature is incorrect for the 'test' package. The signature must be of the form: func Render(string) string (code=5)\n"

	tests := []test{
		{
			desc: "no render function",
			input: `
				package test

				func Random() {}
			`,
		},
		{
			desc: "ignore methods",
			input: `
				package test

				type Test struct{}

				func (t *Test) Render(input int) int {
					return input
				}
			`,
		},

		{
			desc: "wrong parameter type",
			input: `
				package test

				func Render(input int) string {
					return "hello"
				}
			`,
			err: true,
		},
		{
			desc: "wrong return type",
			input: `
				package test

				func Render(input string) int {
					return 9001
				}
			`,
			err: true,
		},
		{
			desc: "too many parameters",
			input: `
				package test

				func Render(input string, extra int) string {
					return input
				}
			`,
			err: true,
		},
		{
			desc: "correct signature",
			input: `
				package test

				func Render(input string) string {
					return input
				}
			`,
			err: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			buf := new(bytes.Buffer)
			io := commands.NewTestIO()
			io.SetErr(commands.WriteNopCloser(buf))

			pkg := typesPkg(t, tc.input)
			hasErr := lintRenderSignature(io, pkg)

			switch {
			case tc.err:
				require.True(t, hasErr)
				require.Equal(t, errMsg, buf.String())
			default:
				require.False(t, hasErr)
				require.Empty(t, buf.String())
			}
		})
	}
}

// helper to take in a file body string and return a types.Package.
// makes plenty of assumptions on the path, pkg name, and files
func typesPkg(t *testing.T, input string) *types.Package {
	t.Helper()

	memPkg := &gnovm.MemPackage{
		Name: "test",
		Path: "gno.land/test",
		Files: []*gnovm.MemFile{
			{
				Name: "main.gno",
				Body: input,
			},
		},
	}

	pkg, err := gno.TypeCheckMemPackageTest(memPkg, &mockMemPkgGetter{name: "test", body: input})
	require.NoError(t, err)

	return pkg
}

// provides a simple impl of MemPackageGetter that only assumes a single file
// of 'main.go' using the underlying name and body of the struct
type mockMemPkgGetter struct {
	name string
	body string
}

func (m *mockMemPkgGetter) GetMemPackage(path string) *gnovm.MemPackage {
	return &gnovm.MemPackage{
		Name: m.name,
		Path: path,
		Files: []*gnovm.MemFile{
			{
				Name: "main.go",
				Body: m.body,
			},
		},
	}
}
