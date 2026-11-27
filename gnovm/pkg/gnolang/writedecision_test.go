package gnolang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
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

// modeFields are the machine's mode (with its allocator's view state).
var modeFields = map[string]bool{
	"Realm": true, "viewOnly": true, "viewS3": true, "viewRealm": true, "curCrossing": true,
	"scratch": true, "viewHoldID": true, "currentRealmID": true,
}

// modeSites are the only functions that use a mode setter or write a mode
// field, each with exactly the uses listed (setter or .field, times). A
// "decided" site applies decideCallStart's answer; the rest start no call.
var modeSites = map[string]struct{ why, uses string }{
	"setRealm":              {"setter: writes as a realm, clearing view state", ".Realm×1 .currentRealmID×2 .scratch×1 .viewHoldID×1 .viewOnly×1 .viewRealm×1 .viewS3×1"},
	"setViewOnly":           {"setter: drops the realm via setRealm(nil)", ".scratch×1 .viewHoldID×1 .viewOnly×1 setRealm×1"},
	"enterReadOnlyMode":     {"setter: view-only plus the read-only realm", ".currentRealmID×1 .scratch×1 .viewRealm×1 .viewS3×1 setViewOnly×1"},
	"enterReadOnly":         {"setter wrapper", "enterReadOnlyMode×1"},
	"enterReadOnlyByS3":     {"setter wrapper", "enterReadOnlyMode×1"},
	"setCrossing":           {"setter: the active crossing call", ".curCrossing×1"},
	"pushFrameCall":         {"decided: applies askCallStart's startCross answer", "setRealm×1"},
	"applyCallStart":        {"decided: applies askCallStart's answer", "enterReadOnlyByS3×1 enterReadOnly×1 setRealm×2 setViewOnly×1"},
	"restoreRealmState":     {"restores the caller's saved mode on return or panic", ".currentRealmID×1 .scratch×1 .viewHoldID×1 .viewOnly×1 .viewRealm×1 .viewS3×1 setRealm×1"},
	"NewMachineWithOptions": {"a new machine starts the tx's first crossing call", ".curCrossing×1"},
	"Release":               {"a released machine is zeroed for the pool", "*×1"},
	"SetActivePackage":      {"tooling: runs as the package's realm, no call", "setRealm×1"},
	"runMemPackage":         {"package init under a throwaway realm, no call", "setRealm×2"},
	"PreprocessFiles":       {"package preprocess under a throwaway realm, no call", "setRealm×2"},
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

// isMachineType reports whether x names Machine or *Machine.
func isMachineType(x ast.Expr) bool {
	if sx, ok := x.(*ast.StarExpr); ok {
		x = sx.X
	}
	id, ok := x.(*ast.Ident)
	return ok && id.Name == "Machine"
}

// checkWriteDecision returns every violation of the single-decision rule.
// Limits (go/ast sees no types or scopes): a Machine reached through a call
// result, a field of another struct, a closure's captured alias, reflect or
// unsafe; Machine{...} literals; aliases are per function, so a shadowing
// inner declaration can hide one.
func checkWriteDecision(fset *token.FileSet, files []*ast.File) []string {
	var bad []string
	report := func(n ast.Node, msg string) { bad = append(bad, fset.Position(n.Pos()).String()+": "+msg) }
	calls := map[string]map[string]bool{} // function -> names it calls
	uses := map[string]map[string]int{}   // site -> setter or .field -> times
	use := func(n ast.Node, name, what string) {
		if _, ok := modeSites[name]; !ok {
			report(n, name+" uses mode "+what)
		}
		if uses[name] == nil {
			uses[name] = map[string]int{}
		}
		uses[name][what]++
	}
	check := func(name string, recv, params *ast.FieldList, body ast.Node) {
		if calls[name] == nil {
			calls[name] = map[string]bool{}
		}
		outside := map[string]bool{}      // receiver and parameters
		machines := map[string]bool{}     // variables holding a Machine or *Machine
		manyMachines := map[string]bool{} // slices or maps of Machines
		addParams := func(fl *ast.FieldList, isOutside bool) {
			if fl == nil {
				return
			}
			for _, f := range fl.List {
				for _, n := range f.Names {
					outside[n.Name] = outside[n.Name] || isOutside
					machines[n.Name] = isMachineType(f.Type)
					switch t := f.Type.(type) {
					case *ast.ArrayType:
						manyMachines[n.Name] = isMachineType(t.Elt)
					case *ast.MapType:
						manyMachines[n.Name] = isMachineType(t.Value)
					}
				}
			}
		}
		addParams(recv, true)
		addParams(params, true)
		// isMachine reports whether x is a Machine variable, maybe via & or *.
		isMachine := func(x ast.Expr) bool {
			for {
				switch e := x.(type) {
				case *ast.UnaryExpr:
					x = e.X
				case *ast.StarExpr:
					x = e.X
				case *ast.ParenExpr:
					x = e.X
				case *ast.SelectorExpr: // m.Alloc holds part of the mode
					if e.Sel.Name != "Alloc" {
						return false
					}
					x = e.X
				case *ast.TypeAssertExpr:
					return isMachineType(e.Type)
				case *ast.Ident:
					return machines[e.Name]
				default:
					return false
				}
			}
		}
		// modeTarget reports whether x writes the mode of a Machine variable.
		modeTarget := func(x ast.Expr) bool {
			if sx, ok := x.(*ast.StarExpr); ok {
				return isMachine(sx.X)
			}
			sx, ok := x.(*ast.SelectorExpr)
			return ok && modeFields[sx.Sel.Name] && machines[rootIdent(sx)]
		}
		ast.Inspect(body, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.CallExpr:
				cn := callName(n)
				calls[name][cn] = true
				if pureDecision[name] && cn != "" && !pureDecision[cn] && !pureCallees[cn] {
					report(n, name+" calls "+cn)
				}
			case *ast.FuncLit:
				addParams(n.Type.Params, false)
			case *ast.RangeStmt:
				if id, ok := n.Value.(*ast.Ident); ok {
					machines[id.Name] = manyMachines[rootIdent(n.X)]
				}
			case *ast.SelectorExpr: // a setter called or taken as a method value
				if modeSetters[n.Sel.Name] {
					use(n, name, n.Sel.Name)
				}
			case *ast.UnaryExpr:
				if sx, ok := n.X.(*ast.SelectorExpr); ok && n.Op == token.AND && modeFields[sx.Sel.Name] {
					use(n, name, "&."+sx.Sel.Name)
				}
			case *ast.ValueSpec:
				for i, id := range n.Names {
					machines[id.Name] = n.Type != nil && isMachineType(n.Type) ||
						len(n.Values) == len(n.Names) && isMachine(n.Values[i])
				}
			case *ast.AssignStmt:
				for i, l := range n.Lhs {
					if id, ok := l.(*ast.Ident); ok && len(n.Rhs) == len(n.Lhs) {
						machines[id.Name] = isMachine(n.Rhs[i]) // an alias of a Machine
					}
					if modeTarget(l) {
						what := "*"
						if sx, ok := l.(*ast.SelectorExpr); ok {
							what = "." + sx.Sel.Name
						}
						use(n, name, what)
					}
					if _, local := l.(*ast.Ident); pureDecision[name] && (!local || outside[rootIdent(l)]) {
						report(n, name+" assigns through "+rootIdent(l))
					}
				}
			case *ast.IncDecStmt:
				if _, local := n.X.(*ast.Ident); pureDecision[name] && (!local || outside[rootIdent(n.X)]) {
					report(n, name+" changes "+rootIdent(n.X))
				}
			}
			return true
		})
	}
	for _, file := range files {
		for _, d := range file.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				if fd.Body != nil {
					check(fd.Name.Name, fd.Recv, fd.Type.Params, fd.Body)
				}
			} else {
				check("<package scope>", nil, nil, d)
			}
		}
	}
	for site, u := range uses {
		var got []string
		for what, k := range u {
			got = append(got, fmt.Sprintf("%s×%d", what, k))
		}
		sort.Strings(got)
		if want := modeSites[site].uses; want != "" && strings.Join(got, " ") != want {
			bad = append(bad, fmt.Sprintf("%s uses %q, want %q", site, strings.Join(got, " "), want))
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
		pre := `package gnolang
func (m *Machine) pushFrameCall() { m.askCallStart(nil) }
func (m *Machine) askCallStart(f *callFacts) { m.decideCallStart(f) }
`
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
			"realm-field": pre + `func (o *Machine) helper(r *Realm) { o.Realm = r }`,
			"extra-in-push": `package gnolang
func (m *Machine) pushFrameCall() { m.setRealm(nil); m.setViewOnly(); m.askCallStart(nil) }
func (m *Machine) askCallStart(f *callFacts) { m.decideCallStart(f) }`,
			"alias-in-decide":       pre + `func (m *Machine) decideCallStart(f *callFacts) { mm := m; mm.Realm = nil }`,
			"local-field-in-decide": pre + `func (m *Machine) byValueReadOnly(f *callFacts) { var cs callStart; cs.byValue = true }`,
			"ptr-alias":             pre + `func (m *Machine) helper() { p := &m.viewOnly; *p = true }`,
			"method-value":          pre + `func (m *Machine) helper() { f := m.setViewOnly; f() }`,
			"package-scope":         pre + `var stray = func(m *Machine) { m.setViewOnly() }`,
			"alloc-state":           pre + `func (m *Machine) helper() { m.Alloc.viewHoldID = PkgID{} }`,
			"listed-extra":          pre + `func (m *Machine) runMemPackage() { m.setRealm(nil); m.setRealm(nil); m.setViewOnly() }`,
			"swapped":               pre + `func (m *Machine) enterReadOnly(r *Realm) { m.setRealm(r) }`,
			"lit-param":             pre + `func helper(m *Machine) { func(x *Machine) { x.Realm = nil }(m) }`,
			"range-var":             pre + `func helper(ms []*Machine) { for _, x := range ms { x.viewOnly = true } }`,
			"alloc-alias":           pre + `func (m *Machine) helper() { a := m.Alloc; a.scratch = 0 }`,
			"whole-struct":          pre + `func (m *Machine) helper(o *Machine) { *m = *o }`,
		}
		for name, src := range stray {
			fs := token.NewFileSet()
			f, err := parser.ParseFile(fs, name+".go", src, 0)
			if err != nil {
				t.Fatal(err)
			}
			v := checkWriteDecision(fs, []*ast.File{f})
			t.Logf("%s: %v", name, v)
			if len(v) == 0 {
				t.Errorf("%s: checker found nothing", name)
			}
		}
	})
}

// A call on a machine with no package reads no caller package unless a rule needs it.
func TestCallStartNoPackage(t *testing.T) {
	st := foreignFuncStore()
	m := NewMachineWithOptions(MachineOptions{PkgPath: "", Store: st, Output: io.Discard})
	defer m.Release()
	pn := NewPackageNode("x", "gno.land/p/x", nil)
	bs := new(BlockStmt)
	bs.InitStaticBlock(bs, pn)
	bs.Define("s", TypedValue{T: StringType, V: StringValue("abc")})
	gx := Preprocess(st, bs, m.MustParseExpr("len(s)")).(Expr)
	if got := m.EvalStatic(bs, gx).String(); got != "(3 int)" {
		t.Fatalf("got %s, want (3 int)", got)
	}
}
