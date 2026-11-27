package diff

import (
	"go/parser"
	"go/scanner"
	"go/token"
	"strconv"
)

// Unsupported reports why src is outside the language subset this fuzzer
// compares. The empty string means the source is in scope (it may still
// fail to type-check). The check uses go/scanner so comments, strings, raw
// strings, and rune literals are handled accurately.
func Unsupported(src string) string {
	if why := unsupportedTokens(src); why != "" {
		return why
	}
	return unsupportedImport(src)
}

type tokItem struct {
	tok token.Token
	lit string
}

func scanTokens(src string) []tokItem {
	fset := token.NewFileSet()
	file := fset.AddFile("main.go", fset.Base(), len(src))
	var s scanner.Scanner
	s.Init(file, []byte(src), nil, 0)
	var toks []tokItem
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		toks = append(toks, tokItem{tok: tok, lit: lit})
	}
	return toks
}

func unsupportedTokens(src string) string {
	toks := scanTokens(src)

	// Track user-declared functions with names matching Go 1.21+ / complex
	// builtins so user-defined helpers (e.g. func min(a, b int) int) stay in scope.
	declaredFunc := map[string]bool{}
	for i := 0; i+1 < len(toks); i++ {
		if toks[i].tok == token.FUNC && toks[i+1].tok == token.IDENT {
			declaredFunc[toks[i+1].lit] = true
		}
	}

	for i, it := range toks {
		switch it.tok {
		case token.CHAN, token.ARROW:
			return "channels are not in the Gno subset"
		case token.SELECT:
			return "select is not in the Gno subset"
		case token.GO:
			return "goroutines are not in the Gno subset"
		case token.IMAG:
			return "complex numbers are not in the Gno subset"
		case token.FUNC, token.TYPE:
			if i+2 < len(toks) && toks[i+1].tok == token.IDENT && toks[i+2].tok == token.LBRACK {
				return "type parameters are not in the Gno subset"
			}
		case token.RANGE:
			if i+1 < len(toks) && toks[i+1].tok == token.INT {
				return "integer range (Go 1.22) is not targeted by this fuzzer"
			}
		case token.IDENT:
			switch it.lit {
			case "unsafe":
				return "unsafe is not in the Gno subset"
			case "complex64", "complex128":
				return "complex numbers are not in the Gno subset"
			case "complex", "real", "imag":
				if isBuiltinCall(toks, i, declaredFunc) {
					return "complex numbers are not in the Gno subset"
				}
			case "min", "max", "clear":
				if isBuiltinCall(toks, i, declaredFunc) {
					return it.lit + " builtin is not in the Gno subset (added after Go 1.17)"
				}
			}
		}
	}
	return ""
}

func isBuiltinCall(toks []tokItem, i int, declaredFunc map[string]bool) bool {
	if i+1 >= len(toks) || toks[i+1].tok != token.LPAREN {
		return false
	}
	if i > 0 {
		switch toks[i-1].tok {
		case token.PERIOD, token.FUNC, token.RPAREN:
			return false
		}
	}
	return !declaredFunc[toks[i].lit]
}

// allowedImport is the stdlib subset Gno ships for production code and whose
// API matches gc closely enough to differential-test. fmt is test-only in
// Gno (it needs reflect), so it stays out.
func allowedImport(path string) bool {
	switch path {
	case "math", "strings", "strconv", "unicode/utf8":
		return true
	default:
		return false
	}
}

func unsupportedImport(src string) string {
	f, _ := parser.ParseFile(token.NewFileSet(), "main.go", src, parser.ImportsOnly)
	if f == nil {
		return ""
	}
	for _, imp := range f.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return "imports could not be parsed"
		}
		if !allowedImport(path) {
			return "imports of " + path + " are out of scope"
		}
	}
	return ""
}
