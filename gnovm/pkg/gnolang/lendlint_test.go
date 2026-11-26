package gnolang_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/store"
)

const (
	lendLintRows   = "../../tests/lendlint_rows.txt"
	lendLintFiles  = "../../tests/files"
	lendLintNotRef = "# flagged but not refused"
)

// The refusal texts of machine.go (readonly access), uverse.go (builtins on
// readonly data) and realm.go (save).
var lendLintRefusals = [][]byte{
	[]byte("is read-only here"),
	[]byte("cannot delete from readonly tainted map"),
	[]byte("cannot copy to readonly tainted slice"),
	[]byte("cannot append to readonly tainted slice"),
	[]byte("was entered without crossing and is read-only"),
	[]byte("an object written by read-only code would be stored"),
}

func lendLintRefuses(src []byte) bool {
	for _, r := range lendLintRefusals {
		if bytes.Contains(src, r) {
			return true
		}
	}
	return false
}

// lendLinter preprocesses one file test and runs the check over its package
// and the realms it imports. LENDLINT_UPDATE=1 rewrites the rows file.
type lendLinter struct {
	base  store.CommitStore
	store gno.Store
}

func newLendLinter(t *testing.T) *lendLinter {
	t.Helper()
	rootDir, err := filepath.Abs("../../../")
	if err != nil {
		t.Fatal(err)
	}
	base, gs := test.StoreWithOptions(rootDir, io.Discard, test.StoreOptions{WithExtern: true, WithExamples: true, Testing: true})
	return &lendLinter{base: base, store: gs}
}

// sites returns "-" or the flagged sites; with own, only the fixture's own.
func (l *lendLinter) sites(fname string, src []byte, own bool) string {
	dirs, err := test.ParseDirectives(bytes.NewReader(src))
	if err != nil {
		return "-"
	}
	pkgPath := dirs.FirstDefault(test.DirectivePkgPath, "main")
	tcw := l.base.CacheWrap()
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: pkgPath,
		Output:  io.Discard,
		Store:   l.store.BeginTransaction(tcw, tcw, nil, nil),
	})
	defer m.Release()
	var out []string
	func() {
		defer func() { _ = recover() }() // a fixture that expects a preprocess error
		fn := m.MustParseFile(fname, string(src))
		pn, _ := m.PreprocessFiles(string(fn.PkgName), pkgPath, &gno.FileSet{Files: []*gno.FileNode{fn}}, false, false)
		pns := []*gno.PackageNode{pn}
		for _, d := range fn.Decls {
			if id, ok := d.(*gno.ImportDecl); ok && !own && gno.IsRealmPath(id.PkgPath) {
				if ipn, ok := m.Store.GetBlockNode(gno.PackageNodeLocation(id.PkgPath)).(*gno.PackageNode); ok {
					pns = append(pns, ipn)
				}
			}
		}
		for _, p := range pns {
			for _, w := range gno.LentCallbackWarnings(m.Store, p) {
				site := fmt.Sprintf("%d:%d", w.Line, w.Column)
				if p != pn {
					site = p.PkgPath + "/" + w.File + ":" + site
				}
				out = append(out, site)
			}
		}
	}()
	if len(out) == 0 {
		return "-"
	}
	return strings.Join(out, ",")
}

// warnings preprocesses src as a realm file and runs the check n times.
func (l *lendLinter) warnings(t *testing.T, src string, n int) [][]gno.LendWarning {
	t.Helper()
	tcw := l.base.CacheWrap()
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: "gno.land/r/x/ll",
		Output:  io.Discard,
		Store:   l.store.BeginTransaction(tcw, tcw, nil, nil),
	})
	defer m.Release()
	fn := m.MustParseFile("ll.gno", src)
	pn, _ := m.PreprocessFiles("ll", "gno.land/r/x/ll", &gno.FileSet{Files: []*gno.FileNode{fn}}, false, false)
	var out [][]gno.LendWarning
	for range n {
		out = append(out, gno.LentCallbackWarnings(m.Store, pn))
	}
	return out
}

// The variable a message names must not depend on map order.
func TestLendLintStable(t *testing.T) {
	const src = `package ll
var a, b int
var saved func()
func g0() { a++ }
func g1() { g0() }
func h()  { b++ }
func f()  { g1(); h() }
func Keep(cur realm) { saved = func() { f() } }
`
	runs := newLendLinter(t).warnings(t, src, 20)
	for i, ws := range runs {
		if len(ws) != 1 || !strings.Contains(ws[0].Msg, "writes a ") {
			t.Fatalf("run %d: got %+v, want one warning naming a", i, ws)
		}
	}
}

// An internal failure leaves one warning saying the check did not finish.
func TestLendLintUnfinished(t *testing.T) {
	pn := gno.NewPackageNode("ll", "gno.land/r/x/ll", &gno.FileSet{})
	pn.FileSet.Files = []*gno.FileNode{nil} // makes the check panic
	ws := gno.LentCallbackWarnings(nil, pn)
	if len(ws) != 1 || ws[0].Line != 0 || ws[0].Msg != "the lent-callback check could not finish for this package" {
		t.Fatalf("got %+v", ws)
	}
}

// lendPanicStore panics on its n-th block-node lookup (n < 0 only counts).
type lendPanicStore struct {
	gno.Store
	calls, n int
}

func (s *lendPanicStore) GetBlockNode(loc gno.Location) gno.BlockNode {
	s.calls++
	if s.calls == s.n {
		panic("lendPanicStore")
	}
	return s.Store.GetBlockNode(loc)
}

// A failure late in the check keeps the warnings already found.
func TestLendLintUnfinishedKeeps(t *testing.T) {
	const src = `package ll
import "gno.land/p/nt/ufmt/v0"
var count int
var saved func()
func Keep(cur realm) { saved = func() { count++ } }
func Later(cur realm) { _ = ufmt.Sprintf("%v", func() { count++ }) }
`
	l := newLendLinter(t)
	tcw := l.base.CacheWrap()
	m := gno.NewMachineWithOptions(gno.MachineOptions{PkgPath: "gno.land/r/x/ll", Output: io.Discard, Store: l.store.BeginTransaction(tcw, tcw, nil, nil)})
	defer m.Release()
	fn := m.MustParseFile("ll.gno", src)
	pn, _ := m.PreprocessFiles("ll", "gno.land/r/x/ll", &gno.FileSet{Files: []*gno.FileNode{fn}}, false, false)
	count := &lendPanicStore{Store: m.Store, n: -1}
	gno.LentCallbackWarnings(count, pn)
	ws := gno.LentCallbackWarnings(&lendPanicStore{Store: m.Store, n: count.calls}, pn)
	last := len(ws) - 1
	if last < 1 || ws[0].Line != 5 || ws[last].Msg != "the lent-callback check could not finish for this package" {
		t.Fatalf("lookups %d, got %+v", count.calls, ws)
	}
}

type lendRow struct {
	name, cat, today, sites, note string
}

func (r lendRow) String() string {
	s := strings.Join([]string{r.name, r.cat, r.today, r.sites}, " ")
	if r.note != "" {
		s += " " + r.note
	}
	return s
}

func parseLendRow(line string) (lendRow, bool) {
	body, note, _ := strings.Cut(line, "#")
	f := strings.Fields(body)
	if len(f) < 2 {
		return lendRow{}, false
	}
	r := lendRow{name: f[0], cat: f[1]}
	if note != "" {
		r.note = "#" + note
	}
	switch {
	case strings.HasPrefix(r.name, "zrealm_") && len(f) == 2: // flagged but not refused
		r.cat, r.sites = "", f[1]
	case len(f) == 4 && (r.cat == "caught" || r.cat == "partly" || r.cat == "missed"):
		r.today, r.sites = f[2], f[3]
	default:
		return lendRow{}, false
	}
	return r, true
}

func TestLendLintRows(t *testing.T) {
	raw, err := os.ReadFile(lendLintRows)
	if err != nil {
		t.Fatal(err)
	}
	l := newLendLinter(t)
	var head []string
	var rows, notRef []lendRow
	inNotRef := false
	for _, line := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
		if r, ok := parseLendRow(line); ok {
			if inNotRef {
				notRef = append(notRef, r)
			} else {
				rows = append(rows, r)
			}
			continue
		}
		if line == lendLintNotRef {
			inNotRef = true
		}
		if !inNotRef && strings.HasPrefix(line, "#") {
			head = append(head, line)
		}
	}
	if len(rows) < 81 {
		t.Fatalf("rows file has %d fixture rows, want the study's 81", len(rows))
	}

	// Per row: today's outcome from the golden, and the flagged sites.
	refusing, flagged, caughtWriting := 0, 0, 0
	got := make([]lendRow, len(rows))
	for i, r := range rows {
		fname := "zrealm_" + r.name + ".gno"
		src, err := os.ReadFile(filepath.Join(lendLintFiles, fname))
		if err != nil {
			t.Fatalf("%s: %v", r.name, err)
		}
		g := r
		g.today = "writes"
		if lendLintRefuses(src) {
			g.today = "refuses"
			refusing++
		}
		g.sites = l.sites(fname, src, false)
		if g.today == "refuses" && g.sites != "-" {
			flagged++
		}
		if r.cat == "caught" && g.today == "writes" {
			caughtWriting++
		}
		if r.cat == "caught" && g.today == "refuses" && g.sites == "-" {
			t.Errorf("%s: a caught row that still refuses has no flagged site", r.name)
		}
		if g.today != r.today || g.sites != r.sites {
			t.Errorf("row %s: got %q, want %q", r.name, g.today+" "+g.sites, r.today+" "+r.sites)
		}
		got[i] = g
	}

	// No flag where nothing is refused, unless listed with a reason. Imported
	// realms are left to the examples run: their goldens need not reach them.
	files, _ := filepath.Glob(filepath.Join(lendLintFiles, "zrealm_*.gno"))
	listed := map[string]lendRow{}
	for _, r := range notRef {
		listed[r.name] = r
	}
	var gotNotRef []lendRow
	for _, f := range files {
		src, _ := os.ReadFile(f)
		if lendLintRefuses(src) {
			continue
		}
		fname := filepath.Base(f)
		sites := l.sites(fname, src, true)
		if sites == "-" {
			continue
		}
		name := strings.TrimSuffix(fname, ".gno")
		r, ok := listed[name]
		switch {
		case !ok:
			t.Errorf("%s: flagged at %s but its golden shows no refusal; fix the check or list it under %q with a reason", name, sites, lendLintNotRef)
		case r.sites != sites:
			t.Errorf("%s: flagged at %s, listed at %s", name, sites, r.sites)
		case r.note == "":
			t.Errorf("%s: listed under %q without a reason", name, lendLintNotRef)
		}
		delete(listed, name)
		gotNotRef = append(gotNotRef, lendRow{name: name, sites: sites, note: r.note})
	}
	for name := range listed {
		t.Errorf("%s: listed under %q but not flagged (or now refused)", name, lendLintNotRef)
	}

	counts := []string{
		fmt.Sprintf("flagged: %d of %d rows that refuse today", flagged, refusing),
		fmt.Sprintf("T69 caught rows now writing: %d", caughtWriting),
	}
	for _, c := range counts {
		if !strings.Contains(string(raw), "\n"+c+"\n") {
			t.Errorf("rows file lacks the count line %q", c)
		}
	}

	if os.Getenv("LENDLINT_UPDATE") != "" {
		var b strings.Builder
		for _, h := range head {
			b.WriteString(h + "\n")
		}
		for _, g := range got {
			b.WriteString(g.String() + "\n")
		}
		b.WriteString(lendLintNotRef + "\n")
		sort.Slice(gotNotRef, func(i, j int) bool { return gotNotRef[i].name < gotNotRef[j].name })
		for _, r := range gotNotRef {
			b.WriteString(r.name + " " + r.sites)
			if r.note != "" {
				b.WriteString(" " + r.note)
			}
			b.WriteString("\n")
		}
		b.WriteString(strings.Join(counts, "\n") + "\n")
		if err := os.WriteFile(lendLintRows, []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
