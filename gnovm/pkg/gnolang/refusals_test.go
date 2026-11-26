package gnolang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
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

var (
	refusalHoleRe  = regexp.MustCompile(`%[sv]`)
	refusalSpaceRe = regexp.MustCompile(`\s+`)
)

// refusalPattern matches s in prose: whitespace may wrap, a hole is any one word.
func refusalPattern(s string) string {
	parts := refusalHoleRe.Split(refusalSpaceRe.ReplaceAllString(s, " "), -1)
	for i, p := range parts {
		parts[i] = strings.ReplaceAll(regexp.QuoteMeta(p), " ", `\s+`)
	}
	return strings.Join(parts, `[^\s|]+?`)
}

// refusalStaleBrackets are brackets refusals no longer print, caught even when
// quoted without their opening words.
var refusalStaleBrackets = []string{
	"call a function of the owning realm",
	"use a method or crossing function)",
	"during a crossing call into that realm)",
	"a /p/ package's object cannot be written after its init)",
	"write it inside a crossing call of",
	"make the function that wrote it a crossing function of",
}

// TestRefusalQuotesCurrent fails when a .md, .txtar or .gno file quotes a
// refusal's opening words followed by a bracket other than its current one, or
// a stale bracket. Known limit: opening words quoted with no bracket pass.
func TestRefusalQuotesCurrent(t *testing.T) {
	type check struct{ open, bracket *regexp.Regexp }
	var checks []check
	for _, r := range refusals {
		i := strings.Index(r.text, " (")
		j := strings.Index(r.text, ")")
		if r.master || i < 0 || j < i {
			continue
		}
		open := strings.TrimPrefix(r.text[:i], "%s")
		checks = append(checks, check{
			regexp.MustCompile(refusalPattern(open) + `\s*\(`),
			regexp.MustCompile(`^` + refusalPattern(r.text[i+1:j+1])),
		})
	}
	var stale []*regexp.Regexp
	for _, b := range refusalStaleBrackets {
		stale = append(stale, regexp.MustCompile(refusalPattern(b)))
	}
	err := filepath.WalkDir("../../..", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if strings.HasPrefix(name, ".") && path != "../../.." || name == "node_modules" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if ext := filepath.Ext(name); d.IsDir() || ext != ".md" && ext != ".txtar" && ext != ".gno" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		doc := string(raw)
		report := func(at int, what string) {
			line := strings.Count(doc[:at], "\n") + 1
			t.Errorf("%s:%d quotes a stale bracket: %.100q", path, line, what)
		}
		for _, c := range checks {
			for _, loc := range c.open.FindAllStringIndex(doc, -1) {
				if rest := doc[loc[1]-1:]; !c.bracket.MatchString(rest) {
					report(loc[0], rest)
				}
			}
		}
		for _, re := range stale {
			for _, loc := range re.FindAllStringIndex(doc, -1) {
				report(loc[0], doc[loc[0]:loc[1]])
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// refusalMasterTexts are master's texts, verbatim: realm code can recover a
// refusal and store it, so a master=true entry must keep its text.
var refusalMasterTexts = map[string]bool{
	"cannot cross-call a non-crossing function %s.%v from %s":   true,
	"cannot cur-call to external realm function %s.%v from %s":  true,
	"cannot cur-call to external realm function %s.%v from %v":  true,
	"cannot append to readonly tainted slice":                   true,
	"cannot copy to readonly tainted slice":                     true,
	"cannot delete from readonly tainted map":                   true,
	"illegal conversion of readonly or externally stored value": true,
	"illegal conversion to external realm type":                 true,
	"cannot mutate %s: package is immutable post-init":          true,
}

func TestRefusalMasterTexts(t *testing.T) {
	n := 0
	for _, r := range refusals {
		if r.master {
			n++
			if !refusalMasterTexts[r.text] {
				t.Errorf("master=true text %q is not master's", r.text)
			}
		}
	}
	if n != len(refusalMasterTexts) {
		t.Errorf("%d master=true entries, want %d", n, len(refusalMasterTexts))
	}
}

// refusalLiteralJoin reports whether n adds words to a refusal: a literal or a
// package-level text name joined by + or as call arguments, or a slice of one.
// Known limits: a local variable holding words, and fmt over a slice of texts.
func refusalLiteralJoin(n ast.Node, names, texts map[string]bool) bool {
	var ops []ast.Expr
	switch x := n.(type) {
	case *ast.SliceExpr:
		id, ok := ast.Unparen(x.X).(*ast.Ident)
		return ok && names[id.Name]
	case *ast.BinaryExpr:
		var flat func(e ast.Expr)
		flat = func(e ast.Expr) {
			if b, ok := ast.Unparen(e).(*ast.BinaryExpr); ok && b.Op == token.ADD {
				flat(b.X)
				flat(b.Y)
				return
			}
			ops = append(ops, ast.Unparen(e))
		}
		flat(x)
	case *ast.CallExpr:
		for _, a := range x.Args {
			ops = append(ops, ast.Unparen(a))
		}
	}
	con, lit := false, false
	for _, e := range ops {
		switch e := e.(type) {
		case *ast.Ident:
			con = con || names[e.Name]
			lit = lit || texts[e.Name]
		case *ast.BasicLit:
			lit = lit || e.Kind == token.STRING
		case *ast.CallExpr: // readonlyAccessPanic returns a refusal
			sel, ok := e.Fun.(*ast.SelectorExpr)
			con = con || ok && sel.Sel.Name == "readonlyAccessPanic"
		}
	}
	return con && lit
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
	files := map[string]*ast.File{}
	texts := map[string]bool{}
	for _, file := range append([]string{"refusals.go"}, refusalRuleFiles...) {
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		files[file] = f
		for _, d := range f.Decls {
			if gd, ok := d.(*ast.GenDecl); ok && gd.Tok != token.IMPORT && gd.Tok != token.TYPE {
				for _, sp := range gd.Specs {
					vs := sp.(*ast.ValueSpec)
					for i, n := range vs.Names {
						if i < len(vs.Values) && !names[n.Name] {
							lit, ok := vs.Values[i].(*ast.BasicLit)
							texts[n.Name] = texts[n.Name] || ok && lit.Kind == token.STRING
						}
					}
				}
			}
		}
	}
	for file, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			if refusalLiteralJoin(n, names, texts) {
				t.Errorf("%s: words added to a refusal text", fset.Position(n.Pos()))
			}
			return true
		})
		if file == "refusals.go" {
			continue
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
