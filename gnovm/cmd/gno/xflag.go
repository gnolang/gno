package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strconv"
	"strings"
)

// patchXVars parses body as Go source and rewrites the string-literal
// initializer of any package-level `var name = "..."` (or
// `var name T = "..."`, including grouped `var ( ... )` blocks)
// declaration whose name matches a key in overrides, similar to
// `go build -ldflags "-X pkg.name=value"`.
//
// Unlike a text-based find/replace, this walks the parsed AST and only
// descends into top-level declarations, so:
//   - a string or raw string literal elsewhere in the file that merely
//     looks like a var declaration (e.g. inside a backtick-quoted
//     template string) is never touched, and
//   - local variables declared inside function bodies are never touched,
//     even if they shadow a package-level name in overrides.
//
// If overrides is empty, body is returned unchanged. If body fails to
// parse, body is also returned unchanged: the caller's own subsequent
// parse of the file (with gno's parser) will surface the real syntax
// error, so this function doesn't need to duplicate that diagnostic.
func patchXVars(fname, body string, overrides map[string]string) string {
	if len(overrides) == 0 {
		return body
	}

	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, fname, body, parser.ParseComments)
	if err != nil {
		return body
	}

	patched := false

	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, name := range valueSpec.Names {
				if i >= len(valueSpec.Values) {
					continue // no initializer to patch, e.g. `var x string`
				}

				value, ok := overrides[name.Name]
				if !ok {
					continue
				}

				lit, ok := valueSpec.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue // not a simple string literal initializer
				}

				lit.Value = strconv.Quote(value)
				patched = true
			}
		}
	}

	if !patched {
		return body
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, astFile); err != nil {
		return body
	}

	return buf.String()
}

// xFlag implements flag.Value, collecting repeated `-X name=value` pairs
// into a map, mirroring the semantics of `go build -ldflags "-X ...=..."`.
type xFlag struct {
	values map[string]string
}

// newXFlag returns an initialized, empty xFlag.
func newXFlag() *xFlag {
	return &xFlag{values: map[string]string{}}
}

// String implements flag.Value.
func (x *xFlag) String() string {
	if x == nil || len(x.values) == 0 {
		return ""
	}

	parts := make([]string, 0, len(x.values))
	for k, v := range x.values {
		parts = append(parts, k+"="+v)
	}

	return strings.Join(parts, ",")
}

// Set implements flag.Value. It parses a single "name=value" pair and
// records it, overwriting any previous value for the same name. It may be
// called multiple times (once per -X flag occurrence on the command line).
func (x *xFlag) Set(s string) error {
	name, value, found := strings.Cut(s, "=")
	if !found || name == "" {
		return fmt.Errorf("invalid -X value %q: expected format name=value", s)
	}

	x.values[name] = value

	return nil
}
