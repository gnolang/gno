package gnolang

import (
	"bytes"
	"errors"
	"go/format"
	"go/parser"
	goscanner "go/scanner"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrecompile(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                      string
		source                    string
		expectedOutput            string
		expectedPreprocessorError error
	}{
		{
			name:           "hello",
			source:         "package foo\nfunc hello() string { return \"world\"}",
			expectedOutput: "package foo\nfunc hello() string { return \"world\"}",
		}, {
			name:           "use-std",
			source:         "package foo\nimport \"std\"\nfunc hello() string { _ = std.Foo\nreturn \"world\"}",
			expectedOutput: "package foo\nimport \"github.com/gnolang/gno/gnovm/stdlibs/stdshim\"\nfunc hello() string { _ = std.Foo\nreturn \"world\"}",
		}, {
			name:           "use-realm",
			source:         "package foo\nimport \"gno.land/r/users\"\nfunc foo()  { _ = users.Register}",
			expectedOutput: "package foo\nimport \"github.com/gnolang/gno/examples/gno.land/r/users\"\nfunc foo() { _ = users.Register}",
		}, {
			name:           "use-avl",
			source:         "package foo\nimport \"gno.land/p/demo/avl\"\nfunc foo()  { _ = avl.Tree}",
			expectedOutput: "package foo\nimport \"github.com/gnolang/gno/examples/gno.land/p/demo/avl\"\nfunc foo() { _ = avl.Tree}",
		}, {
			name:           "use-named-std",
			source:         "package foo\nimport bar \"std\"\nfunc hello() string { _ = bar.Foo\nreturn \"world\"}",
			expectedOutput: "package foo\nimport bar \"github.com/gnolang/gno/gnovm/stdlibs/stdshim\"\nfunc hello() string { _ = bar.Foo\nreturn \"world\"}",
		}, {
			name:                      "blacklisted-package",
			source:                    "package foo\nimport \"reflect\"\nfunc foo() { _ = reflect.ValueOf}",
			expectedPreprocessorError: errors.New(`import "reflect" is not in the whitelist`),
		}, {
			name:           "whitelisted-package",
			source:         "package foo\nimport \"regexp\"\nfunc foo() { _ = regexp.MatchString}",
			expectedOutput: "package foo\nimport \"regexp\"\nfunc foo() { _ = regexp.MatchString}",
		},
		// multiple files
		// syntax error
		// unknown realm?
		// blacklist
		// etc
	}
	for _, c := range cases {
		c := c // scopelint
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// parse gno
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "foo.go", c.source, parser.ParseComments)
			assert.NoError(t, err)

			// call preprocessor
			transformed, err := precompileAST(fset, f, true)
			if c.expectedPreprocessorError == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, err, c.expectedPreprocessorError)
			}

			// generate go
			var buf bytes.Buffer
			err = format.Node(&buf, fset, transformed)
			assert.NoError(t, err)
			got := buf.Bytes()

			// check output
			if c.expectedOutput != "" {
				expect, err := format.Source([]byte(c.expectedOutput))
				if !bytes.Equal(expect, got) {
					t.Logf("got:\n%s", got)
					t.Logf("expect:\n%s", expect)
					t.Fatal("mismatch")
				}
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseGoBuildErrors(t *testing.T) {
	tests := []struct {
		name          string
		output        string
		expectedError error
	}{
		{
			name:          "empty output",
			output:        "",
			expectedError: nil,
		},
		{
			name:          "random output",
			output:        "xxx",
			expectedError: nil,
		},
		{
			name: "some errors",
			output: `xxx
./main.gno.gen.go:6:2: nasty error
./pkg/file.gno.gen.go:60:20: ugly error`,
			expectedError: goscanner.ErrorList{
				&goscanner.Error{
					Pos: token.Position{
						Filename: "./main.gno",
						Line:     1,
						Column:   1,
					},
					Msg: "nasty error",
				},
				&goscanner.Error{
					Pos: token.Position{
						Filename: "./pkg/file.gno",
						Line:     55,
						Column:   19,
					},
					Msg: "ugly error",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseGoBuildErrors(tt.output)

			assert.Equal(t, tt.expectedError, err)
		})
	}
}
