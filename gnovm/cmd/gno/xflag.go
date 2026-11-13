package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"
)

// xOverride is a single -X override: the value of variable Name in package
// PkgPath should be replaced with Value.
type xOverride struct {
	PkgPath string
	Name    string
	Value   string
}

// xFlag implements flag.Value, collecting repeated `-X importpath.name=value`
// pairs, mirroring the required qualified form of
// `go build -ldflags "-X importpath.name=value"` (see cmd/link docs).
// Unlike an unqualified `-X name=value`, this ties an override to one
// specific package, so it can't accidentally rewrite a same-named var in
// every package under test.
type xFlag struct {
	overrides []xOverride
}

// newXFlag returns an initialized, empty xFlag.
func newXFlag() *xFlag {
	return &xFlag{}
}

// String implements flag.Value.
func (x *xFlag) String() string {
	if x == nil || len(x.overrides) == 0 {
		return ""
	}

	parts := make([]string, 0, len(x.overrides))
	for _, o := range x.overrides {
		parts = append(parts, o.PkgPath+"."+o.Name+"="+o.Value)
	}

	return strings.Join(parts, ",")
}

// Set implements flag.Value. It parses a single "importpath.name=value"
// triple and records it; it may be called multiple times (once per -X
// flag occurrence on the command line).
func (x *xFlag) Set(s string) error {
	eq := strings.IndexByte(s, '=')
	if eq < 0 {
		return fmt.Errorf("invalid -X value %q: expected format importpath.name=value", s)
	}

	qualified, value := s[:eq], s[eq+1:]

	dot := strings.LastIndexByte(qualified, '.')
	if dot < 0 || dot == len(qualified)-1 {
		return fmt.Errorf(
			"invalid -X value %q: expected format importpath.name=value (e.g. -X main.Version=1.2.3)",
			s,
		)
	}

	pkgPath, name := qualified[:dot], qualified[dot+1:]
	if pkgPath == "" {
		return fmt.Errorf("invalid -X value %q: empty package path", s)
	}

	x.overrides = append(x.overrides, xOverride{PkgPath: pkgPath, Name: name, Value: value})

	return nil
}

// forPackage returns the name->value overrides that apply to a package,
// identified by its resolved pkgPath and its declared package name
// pkgName. A package is matched either by its full pkgPath, or -- for the
// main package specifically -- by the literal "main", mirroring
// `go build -X main.Var=...`'s own convention of addressing the main
// package by name rather than its (often irrelevant) import path.
func (x *xFlag) forPackage(pkgPath, pkgName string) map[string]string {
	if len(x.overrides) == 0 {
		return nil
	}

	overrides := make(map[string]string)
	for _, o := range x.overrides {
		if o.PkgPath == pkgPath || (pkgName == "main" && o.PkgPath == "main") {
			overrides[o.Name] = o.Value
		}
	}

	return overrides
}

// xPatchResult reports which of the requested overrides were actually
// applied to a package's source, split out by why the rest weren't:
// NotFound names don't correspond to any top-level declaration at all;
// WrongKind names do, but aren't a plain string var (they're a const, a
// var of a different type, or a var with no literal initializer) -- go
// build -X's own error for this case ("cannot set with -X: not a var of
// type string") is the model here.
type xPatchResult struct {
	Matched   []string
	NotFound  []string
	WrongKind []string
}

// Unmatched reports whether any requested override was not applied.
func (r xPatchResult) Unmatched() bool {
	return len(r.NotFound) > 0 || len(r.WrongKind) > 0
}

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
// Rather than reprinting the whole file (which would reformat everything,
// not just the patched literals), each match is spliced into the
// original bytes at the literal's own offsets, leaving the rest of the
// file byte-for-byte untouched. If a matched literal spanned multiple
// source lines (a multi-line raw string) and its replacement doesn't, the
// splice is padded with blank lines so that line numbers for everything
// after it in the file are unaffected -- a later panic or type error still
// points at the same line it would without -X.
//
// If overrides is empty, body is returned unchanged with a zero-value
// result. If body fails to parse, body is also returned unchanged (with a
// zero-value result): the caller's own subsequent parse of the file (with
// gno's parser) will surface the real syntax error, so this function
// doesn't need to duplicate that diagnostic.
func patchXVars(fname, body string, overrides map[string]string) (string, xPatchResult) {
	var result xPatchResult

	if len(overrides) == 0 {
		return body, result
	}

	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, fname, body, parser.ParseComments)
	if err != nil {
		return body, result
	}

	type span struct {
		start, end int
		repl       string
	}

	var spans []span

	matched := make(map[string]bool, len(overrides))
	wrongKind := make(map[string]bool)

	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || (genDecl.Tok != token.VAR && genDecl.Tok != token.CONST) {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, name := range valueSpec.Names {
				value, ok := overrides[name.Name]
				if !ok {
					continue
				}

				lit, ok := literalInitializer(genDecl, valueSpec, i)
				if !ok {
					wrongKind[name.Name] = true
					continue
				}

				matched[name.Name] = true

				start := fset.Position(lit.Pos()).Offset
				end := fset.Position(lit.End()).Offset
				newText := strconv.Quote(value)

				// Preserve line numbers for the rest of the file: pad
				// with as many newlines as the original span had, since
				// the replacement (a double-quoted string) never
				// contains one itself.
				if n := strings.Count(body[start:end], "\n"); n > 0 {
					newText += strings.Repeat("\n", n)
				}

				spans = append(spans, span{start: start, end: end, repl: newText})
			}
		}
	}

	var matchedNames []string
	for name := range overrides {
		switch {
		case matched[name]:
			matchedNames = append(matchedNames, name)
		case wrongKind[name]:
			result.WrongKind = append(result.WrongKind, name)
		default:
			result.NotFound = append(result.NotFound, name)
		}
	}
	sort.Strings(matchedNames)
	sort.Strings(result.NotFound)
	sort.Strings(result.WrongKind)
	result.Matched = matchedNames

	if len(spans) == 0 {
		return body, result
	}

	sort.Slice(spans, func(i, j int) bool { return spans[i].start < spans[j].start })

	var buf strings.Builder

	pos := 0
	for _, sp := range spans {
		buf.WriteString(body[pos:sp.start])
		buf.WriteString(sp.repl)
		pos = sp.end
	}
	buf.WriteString(body[pos:])

	return buf.String(), result
}

// literalInitializer returns the string-literal initializer of the
// name-th name in valueSpec, and whether it's eligible for -X patching:
// genDecl must be a var (not const) declaration, i must have a
// corresponding initializer expression, and that expression must be a
// plain (non-typed-conversion) string literal.
func literalInitializer(genDecl *ast.GenDecl, valueSpec *ast.ValueSpec, i int) (*ast.BasicLit, bool) {
	if genDecl.Tok != token.VAR {
		return nil, false // const: go build -X also refuses this.
	}

	if i >= len(valueSpec.Values) {
		return nil, false // e.g. `var x string` with no initializer.
	}

	lit, ok := valueSpec.Values[i].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return nil, false // not a plain string literal initializer.
	}

	return lit, true
}
