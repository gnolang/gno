package gnolang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// refusalRuleFiles raise the write rule's refusals, with refusals.go's texts only.
var refusalRuleFiles = []string{
	"machine.go", "op_assign.go", "realm.go", "callstart.go", "uverse.go",
	"op_expressions.go", "crossingfn.go", "lend.go", "birthrule.go",
}

var (
	// a bracketed "do this instead": at least two words, not a format verb
	refusalFixRe = regexp.MustCompile(`\([a-z][^()%]*\s[^()]*\)`)
	// words only a refusal of the write rule uses
	refusalWordRe = regexp.MustCompile(`(?i)read-?only|tainted|cross-call|cur-call|immutable post-init|external realm|another realm`)
)

func TestRefusalTextsNameFix(t *testing.T) {
	for _, text := range refusalTexts {
		if !refusalFixRe.MatchString(text) {
			t.Errorf("refusal text names no fix in brackets: %q", text)
		}
	}
	fset := token.NewFileSet()
	for _, name := range refusalRuleFiles {
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			s, err := strconv.Unquote(lit.Value)
			if err != nil || !refusalWordRe.MatchString(s) || strings.HasPrefix(s, "invariant violation:") {
				return true
			}
			t.Errorf("%s: refusal text not a constant of refusals.go: %q", fset.Position(lit.Pos()), s)
			return true
		})
	}
	// Every constant of refusals.go is in refusalTexts.
	f, err := parser.ParseFile(fset, "refusals.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	listed := map[string]bool{}
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, sp := range gd.Specs {
			vs, ok := sp.(*ast.ValueSpec)
			if !ok {
				continue
			}
			if gd.Tok == token.VAR && vs.Names[0].Name == "refusalTexts" {
				for _, e := range vs.Values[0].(*ast.CompositeLit).Elts {
					listed[e.(*ast.Ident).Name] = true
				}
			}
		}
	}
	for _, d := range f.Decls {
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.CONST {
			for _, sp := range gd.Specs {
				for _, n := range sp.(*ast.ValueSpec).Names {
					if !listed[n.Name] {
						t.Errorf("refusals.go: %s is not in refusalTexts", n.Name)
					}
				}
			}
		}
	}
}
