package gnolang

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// initCycleErrors preprocesses src in accumulate mode and returns the
// located initialization-cycle messages, and the gas the walk charged.
func initCycleErrors(t *testing.T, src string) ([]string, int64) {
	t.Helper()
	m := NewMachine("main", nil)
	defer m.Release()
	fn, err := m.ParseFile("main.gno", src)
	if err != nil {
		t.Fatal(err)
	}
	fset := &FileSet{Files: []*FileNode{fn}}
	pn := NewPackageNode("main", "main", fset)
	if errs := PreprocessFilesCollect(m.Store, pn, fset); len(errs) > 0 {
		var msgs []string
		for _, e := range errs {
			msgs = append(msgs, e.Unwrap().Error())
		}
		gm := types.NewGasMeter(1 << 62)
		checkInitCycles(gm, pn, fset)
		return msgs, gm.GasConsumed()
	}
	return nil, 0
}

func TestInitCycleReportedAtCycleMember(t *testing.T) {
	// b depends on a cycle it is not part of: the cycle is reported once,
	// at c, as Go does, and b gets no error.
	msgs, _ := initCycleErrors(t, `package main

var b = h()
var c = T{}.m()

func h() int { return c }

type T struct{}

func (T) m() int { return c }

func main() {}
`)
	want := []string{"main/main.gno:4:5-16: initialization cycle: c refers to T.m refers to c"}
	if fmt.Sprint(msgs) != fmt.Sprint(want) {
		t.Fatalf("got %v, want %v", msgs, want)
	}
}

func TestInitCycleOnePerComponent(t *testing.T) {
	msgs, _ := initCycleErrors(t, `package main

var a = f()
var b = h()
var c = T{}.m()
var d = func() int { return d }()
var e = b + a

func f() int { return a }
func h() int { return c + e }

type T struct{}

func (T) m() int { return d + c }

func main() {}
`)
	want := []string{
		"main/main.gno:3:5-12: initialization cycle: a refers to f refers to a",
		"main/main.gno:4:5-12: initialization cycle: b refers to h refers to e refers to b",
		"main/main.gno:5:5-16: initialization cycle: c refers to T.m refers to c",
		"main/main.gno:6:5-34: initialization cycle: d refers to d",
	}
	if fmt.Sprint(msgs) != fmt.Sprint(want) {
		t.Fatalf("got\n%s\nwant\n%s", strings.Join(msgs, "\n"), strings.Join(want, "\n"))
	}
}

func TestInitCycleStricterThanRuntimeOrdering(t *testing.T) {
	// runFileDecls' ordering walk cached V as cycle-free once A reached d;
	// the component walk sees V <-> d.
	msgs, _ := initCycleErrors(t, "package main\n\nvar A = d()\nvar V = d()\n\nfunc d() int { return V }\n\nfunc main() {}\n")
	if len(msgs) != 1 || !strings.HasSuffix(msgs[0], "initialization cycle: V refers to d refers to V") {
		t.Fatalf("got %v", msgs)
	}
}

func initCycleChain(n int) string {
	var b strings.Builder
	b.WriteString("package main\n")
	for i := range n {
		fmt.Fprintf(&b, "var v%d = f%d()\nfunc f%d() int { return v%d }\n", i, i, i, i+1)
	}
	fmt.Fprintf(&b, "var v%d = f0()\nfunc main() {}\n", n)
	return b.String()
}

func TestInitCycleWalkIsLinear(t *testing.T) {
	// A chain of n variables through n functions closing on v1: one
	// component, walked once. A per-root walk would visit O(n^2) nodes.
	const n = 400
	msgs, gas := initCycleErrors(t, initCycleChain(n))
	// v0 is not on the cycle (f0 returns v1): the report starts at v1.
	if len(msgs) != 1 || !strings.Contains(msgs[0], "initialization cycle: v1 refers to f1") {
		t.Fatalf("got %v", msgs)
	}
	// 2n+1 declarations, 2n+1 edges, each visited by Tarjan's pass and
	// by the walk that names the cycle: 4 charges per node and per edge
	// at most.
	if limit := int64(4*(2*(2*n+1))) * OpCPUSlopeInitCycleVisit; gas > limit {
		t.Fatalf("gas %d exceeds the linear bound %d", gas, limit)
	}
}

func TestInitCycleOverrideRoots(t *testing.T) {
	run := func(first, last string) (msg string) {
		m := NewMachine("", nil)
		defer m.Release()
		fset := &FileSet{Files: []*FileNode{m.MustParseFile("first.gno", first), m.MustParseFile("last.gno", last)}}
		defer func() {
			if r := recover(); r != nil {
				msg = r.(*PreprocessError).Unwrap().Error()
			}
		}()
		m.PreprocessFiles("override", "test/override", fset, false, true)
		return ""
	}
	// A shadowed declaration is neither a root nor a resolution target.
	if msg := run("package override\nvar x, a = a, x\n", "package override\nvar a = 1\n"); msg != "" {
		t.Fatalf("false cycle: %s", msg)
	}
	if msg := run("package override\nvar a = g()\nfunc g() int { return a }\n", "package override\nvar a = 1\n"); msg != "" {
		t.Fatalf("shadowed self-cycle reported: %s", msg)
	}
	// A declaration that still owns one of its names is a root.
	msg := run("package override\nvar a, b = two()\nfunc two() (int, int) { return b, 1 }\n", "package override\nvar a = 2\n")
	if !strings.Contains(msg, "initialization cycle: a refers to two refers to b") {
		t.Fatalf("live b not walked: %q", msg)
	}
}

func BenchmarkInitCycleWalk(b *testing.B) {
	m := NewMachine("main", nil)
	defer m.Release()
	fn := m.MustParseFile("main.gno", initCycleChain(2000))
	fset := &FileSet{Files: []*FileNode{fn}}
	pn := NewPackageNode("main", "main", fset)
	PreprocessFilesCollect(m.Store, pn, fset)
	gm := types.NewGasMeter(1 << 62)
	b.ResetTimer()
	for range b.N {
		checkInitCycles(gm, pn, fset)
	}
	b.ReportMetric(float64(gm.GasConsumed())/float64(b.N)/OpCPUSlopeInitCycleVisit, "visits/op")
}
