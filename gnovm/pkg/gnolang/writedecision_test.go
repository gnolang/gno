package gnolang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// modeSetters change whether running code may write.
var modeSetters = map[string]bool{
	"setRealm": true, "setViewOnly": true, "enterReadOnly": true,
	"enterReadOnlyByS3": true, "enterReadOnlyMode": true,
}

// modeFields are the machine's mode; only a setter site assigns them.
var modeFields = map[string]bool{"viewOnly": true, "viewS3": true, "viewRealm": true}

// modeSetterSites are the only functions that may call a mode setter. A
// "decided" entry applies decideCallStart's answer; the rest are not a call
// start deciding its mode.
var modeSetterSites = map[string]string{
	"setRealm":          "setter: writes as a realm, clearing view state",
	"setViewOnly":       "setter: drops the realm via setRealm(nil)",
	"enterReadOnlyMode": "setter: view-only plus the read-only realm",
	"enterReadOnly":     "setter wrapper",
	"enterReadOnlyByS3": "setter wrapper",
	"pushFrameCall":     "decided: applies askCallStart's startCross answer",
	"applyCallStart":    "decided: applies askCallStart's answer",
	"restoreRealmState": "restores the caller's saved mode on return or panic",
	"SetActivePackage":  "tooling: runs as the package's realm, no call",
	"runMemPackage":     "package init under a throwaway realm, no call",
	"PreprocessFiles":   "package preprocess under a throwaway realm, no call",
}

// pureDecision are decideCallStart and the helpers it reads through, and
// pureCallees every function they may call: none of them changes anything.
var (
	pureDecision = map[string]bool{
		"decideCallStart": true, "byValueReadOnly": true, "lendLiveAs": true,
		"testRunnerCall": true, "calledByValue": true, "readOnlyOf": true,
	}
	pureCallees = map[string]bool{
		"IsCrossing": true, "GetRealm": true, "IsRealmPkg": true, "GetObjectInfo": true,
		"GetAttribute": true, "IsZero": true, "IsMethod": true, "len": true,
	}
)

func callName(c *ast.CallExpr) string {
	switch fx := c.Fun.(type) {
	case *ast.Ident:
		return fx.Name
	case *ast.SelectorExpr:
		return fx.Sel.Name
	}
	return ""
}

// rootIdent is the variable an assignment target writes through.
func rootIdent(x ast.Expr) string {
	for {
		switch e := x.(type) {
		case *ast.Ident:
			return e.Name
		case *ast.SelectorExpr:
			x = e.X
		case *ast.IndexExpr:
			x = e.X
		case *ast.StarExpr:
			x = e.X
		case *ast.ParenExpr:
			x = e.X
		default:
			return ""
		}
	}
}

// checkWriteDecision returns every violation of the single-decision rule.
func checkWriteDecision(fset *token.FileSet, files []*ast.File) []string {
	var bad []string
	calls := map[string]map[string]bool{} // function -> names it calls
	for _, file := range files {
		for _, d := range file.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			name := fd.Name.Name
			if calls[name] == nil {
				calls[name] = map[string]bool{}
			}
			outside := map[string]bool{} // receiver and parameters
			if fd.Recv != nil {
				for _, f := range fd.Recv.List {
					for _, n := range f.Names {
						outside[n.Name] = true
					}
				}
			}
			for _, f := range fd.Type.Params.List {
				for _, n := range f.Names {
					outside[n.Name] = true
				}
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				switch n := n.(type) {
				case *ast.CallExpr:
					cn := callName(n)
					calls[name][cn] = true
					if modeSetters[cn] && modeSetterSites[name] == "" {
						bad = append(bad, fset.Position(n.Pos()).String()+": "+name+" calls mode setter "+cn)
					}
					if pureDecision[name] && cn != "" && !pureDecision[cn] && !pureCallees[cn] {
						bad = append(bad, fset.Position(n.Pos()).String()+": "+name+" calls "+cn)
					}
				case *ast.AssignStmt:
					for _, l := range n.Lhs {
						if sx, ok := l.(*ast.SelectorExpr); ok && modeFields[sx.Sel.Name] && modeSetterSites[name] == "" {
							bad = append(bad, fset.Position(n.Pos()).String()+": "+name+" assigns mode field "+sx.Sel.Name)
						}
					}
					if pureDecision[name] {
						for _, l := range n.Lhs {
							if r := rootIdent(l); outside[r] || r == "" {
								bad = append(bad, fset.Position(n.Pos()).String()+": "+name+" assigns through "+r)
							}
						}
					}
				case *ast.IncDecStmt:
					if pureDecision[name] && outside[rootIdent(n.X)] {
						bad = append(bad, fset.Position(n.Pos()).String()+": "+name+" changes "+rootIdent(n.X))
					}
				}
				return true
			})
		}
	}
	// A "decided" site takes its answer from decideCallStart.
	if !calls["pushFrameCall"]["askCallStart"] || !calls["askCallStart"]["decideCallStart"] {
		bad = append(bad, "pushFrameCall does not take its answer from decideCallStart")
	}
	for fn, cs := range calls {
		if cs["applyCallStart"] && fn != "pushFrameCall" {
			bad = append(bad, fn+" applies a call start it did not decide")
		}
	}
	sort.Strings(bad)
	return bad
}

func TestWriteDecisionSingleSite(t *testing.T) {
	fset := token.NewFileSet()
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	var files []*ast.File
	for _, p := range paths {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		src, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		f, err := parser.ParseFile(fset, p, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
	}
	for _, v := range checkWriteDecision(fset, files) {
		t.Error(v)
	}

	t.Run("stray", func(t *testing.T) {
		stray := map[string]string{
			"setter": `package gnolang
func (m *Machine) pushFrameCall() { m.askCallStart(nil) }
func (m *Machine) askCallStart(f *callFacts) { m.decideCallStart(f) }
func (m *Machine) helper() { m.setViewOnly() }`,
			"impure": `package gnolang
func (m *Machine) pushFrameCall() { m.askCallStart(nil) }
func (m *Machine) askCallStart(f *callFacts) { m.decideCallStart(f) }
func (m *Machine) decideCallStart(f *callFacts) { m.viewOnly = true }`,
			"field": `package gnolang
func (m *Machine) pushFrameCall() { m.askCallStart(nil) }
func (m *Machine) askCallStart(f *callFacts) { m.decideCallStart(f) }
func (m *Machine) helper() { m.viewS3 = false }`,
			"undecided": `package gnolang
func (m *Machine) pushFrameCall() {}
func (m *Machine) askCallStart(f *callFacts) { m.decideCallStart(f) }`,
		}
		for name, src := range stray {
			fs := token.NewFileSet()
			f, err := parser.ParseFile(fs, name+".go", src, 0)
			if err != nil {
				t.Fatal(err)
			}
			if len(checkWriteDecision(fs, []*ast.File{f})) == 0 {
				t.Errorf("%s: checker found nothing", name)
			}
		}
	})
}
