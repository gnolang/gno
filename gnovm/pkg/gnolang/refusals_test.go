package gnolang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// refusalRuleFiles raise the write rule's refusals, with refusals.go's texts only.
var refusalRuleFiles = []string{
	"machine.go", "op_assign.go", "realm.go", "callstart.go", "uverse.go",
	"op_expressions.go", "crossingfn.go", "lend.go", "birthrule.go", "preprocess.go",
}

// refusalOtherRules are texts of the scanned files that the word check would
// match but that belong to the rule on passing cur, not the write rule.
var refusalOtherRules = map[string]bool{
	"only the `cur` argument of a containing crossing function maybe passed by cross-call": true,
}

var (
	// a bracketed "do this instead": at least two words, not a format verb
	refusalFixRe = regexp.MustCompile(`\([a-z](?:[^()%]|\([^()]*\))*\s(?:[^()]|\([^()]*\))*\)`)
	// words only a refusal of the write rule uses
	refusalWordRe = regexp.MustCompile(`(?i)read-?only|tainted|cross-call|cur-call|immutable post-init|external realm|another realm`)
)

func TestRefusalTextsNameFix(t *testing.T) {
	for _, r := range refusals {
		if strings.TrimSpace(r.fix) == "" || strings.TrimSpace(r.when) == "" {
			t.Errorf("refusal has no when or what to do: %q", r.text)
		}
		if !r.master && !refusalFixRe.MatchString(r.text) {
			t.Errorf("refusal text new on this branch names no fix in brackets: %q", r.text)
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
			if err != nil || !refusalWordRe.MatchString(s) || strings.HasPrefix(s, "invariant violation:") || refusalOtherRules[s] {
				return true
			}
			t.Errorf("%s: refusal text not a constant of refusals.go: %q", fset.Position(lit.Pos()), s)
			return true
		})
	}
	// Every constant of refusals.go is in refusals.
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
			if gd.Tok == token.VAR && vs.Names[0].Name == "refusals" {
				for _, e := range vs.Values[0].(*ast.CompositeLit).Elts {
					listed[e.(*ast.CompositeLit).Elts[0].(*ast.Ident).Name] = true
				}
			}
		}
	}
	for _, d := range f.Decls {
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.CONST {
			for _, sp := range gd.Specs {
				for _, n := range sp.(*ast.ValueSpec).Names {
					if !listed[n.Name] {
						t.Errorf("refusals.go: %s is not in refusals", n.Name)
					}
				}
			}
		}
	}
}

// refusalDocRow is the docs table row for r: the text with its run-time parts
// shown as "…", when it fires, and what to do.
func refusalDocRow(r refusal) string {
	hole := strings.NewReplacer("%s.%v", "…", "%s", "…", "%v", "…")
	form := hole.Replace(r.text)
	if strings.HasPrefix(form, "'s ") {
		form = "…" + form
	}
	if strings.HasSuffix(form, ": ") {
		form += "…"
	}
	return "| `" + form + "` | " + r.when + " | " + r.fix + " |"
}

func TestRefusalTextsInDocs(t *testing.T) {
	raw, err := os.ReadFile("../../../docs/resources/gno-interrealm.md")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(raw)
	const head = "| Refusal | When | What to do |\n|---|---|---|\n"
	var want strings.Builder
	for _, r := range refusals {
		want.WriteString(refusalDocRow(r) + "\n")
	}
	// The table on the page must be exactly the generated one, in both directions.
	i := strings.Index(doc, head)
	got := ""
	if i >= 0 {
		got = doc[i+len(head):]
		if j := strings.Index(got, "\n\n"); j >= 0 {
			got = got[:j+1]
		}
	}
	if got != want.String() {
		t.Errorf("gno-interrealm.md's refusal table differs from refusals; generated:\n%s%s", head, want.String())
	}
}

// refusalSites are the functions that raise a refusal of the write rule, keyed
// file:function (a uverse builtin by its name), with how many refusal constants
// or refuseWrite calls each names. Swapping one for a raw string drops the count.
var refusalSites = map[string]int{
	"machine.go:pushFrameCall":        2,
	"machine.go:PopAsPointer":         1,
	"machine.go:resolvePointer":       1,
	"machine.go:readonlyAccessPanic":  4,
	"op_assign.go:doOpAssign":         1,
	"realm.go:DidUpdate":              2,
	"realm.go:assertNoUncrossedWrite": 1,
	"op_expressions.go:convertValue":  2,
	"uverse.go:append":                1,
	"uverse.go:copy":                  2,
	"uverse.go:delete":                1,
	"preprocess.go:preprocess1":       1,
}

func TestRefusalSites(t *testing.T) {
	names := map[string]bool{"refuseWrite": true}
	f, err := parser.ParseFile(token.NewFileSet(), "refusals.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range f.Decls {
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.CONST {
			for _, sp := range gd.Specs {
				for _, n := range sp.(*ast.ValueSpec).Names {
					names[n.Name] = true
				}
			}
		}
	}
	got := map[string]int{}
	fset := token.NewFileSet()
	for _, file := range refusalRuleFiles {
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			key := file + ":" + fd.Name.Name
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if c, ok := n.(*ast.CallExpr); ok && file == "uverse.go" && len(c.Args) > 0 {
					if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "defNative" {
						if lit, ok := c.Args[0].(*ast.BasicLit); ok {
							name, _ := strconv.Unquote(lit.Value)
							ast.Inspect(c, func(n ast.Node) bool {
								if id, ok := n.(*ast.Ident); ok && names[id.Name] {
									got[file+":"+name]++
								}
								return true
							})
							return false
						}
					}
				}
				if id, ok := n.(*ast.Ident); ok && names[id.Name] {
					got[key]++
				}
				return true
			})
		}
	}
	for k, want := range refusalSites {
		if got[k] != want {
			t.Errorf("%s names %d refusal constants or refuseWrite calls, want %d: a refusal raised with a raw string?", k, got[k], want)
		}
	}
	for k, n := range got {
		if _, ok := refusalSites[k]; !ok {
			t.Errorf("%s raises %d refusals but is not in refusalSites", k, n)
		}
	}
}
