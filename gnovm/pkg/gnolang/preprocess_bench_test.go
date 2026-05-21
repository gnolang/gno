package gnolang

import (
	"fmt"
	"strings"
	"testing"
)

// Benchmark sources used to measure Preprocess + post-preprocess coda
// passes. Each source is designed to exercise the coda passes meaningfully:
//   - Top-level var declarations (hits codaHeapDefinesByUse / DemoteDefines
//     ValueDecl paths, i.e. Blocker A territory).
//   - Closures capturing outer vars (hits the heap-escape analysis loop).
//   - Package-scope name references (hits codaPackageSelectors rewrites,
//     i.e. Blocker B territory).
var benchCodaSources = map[string]string{
	"small":  benchSrcSmall,
	"medium": benchSrcMedium(),
	"large":  benchSrcLarge(),
}

const benchSrcSmall = `package bench

var counter int
var greeting = "hello"

func emit(s string) string {
	counter++
	return greeting + s
}

func makeAdder(base int) func(int) int {
	return func(x int) int {
		return base + x + counter
	}
}

func main() {
	add := makeAdder(10)
	_ = add(1) + add(2)
	_ = emit("!")
}
`

// benchSrcMedium generates ~30 top-level vars, ~15 funcs with closure captures,
// and a handful of cross-func references.
func benchSrcMedium() string {
	var b strings.Builder
	b.WriteString("package bench\n\n")
	// Top-level vars — exercises ValueDecl heap-use path.
	for i := 0; i < 30; i++ {
		fmt.Fprintf(&b, "var v%d = %d\n", i, i)
	}
	b.WriteString("\nvar events []string\n")
	b.WriteString(`func emit(s string) string { events = append(events, s); return s }` + "\n\n")
	// Funcs with closures capturing package vars.
	for i := 0; i < 15; i++ {
		fmt.Fprintf(&b, `
func f%d(x int) func() int {
	local := x + v%d
	return func() int {
		return local + v%d + len(events)
	}
}
`, i, i, (i+1)%30)
	}
	b.WriteString("func main() {\n")
	for i := 0; i < 15; i++ {
		fmt.Fprintf(&b, "\t_ = f%d(%d)()\n", i, i)
	}
	b.WriteString(`_ = emit("done")` + "\n")
	b.WriteString("}\n")
	return b.String()
}

// benchSrcLarge scales up medium by ~5x.
func benchSrcLarge() string {
	var b strings.Builder
	b.WriteString("package bench\n\n")
	for i := 0; i < 150; i++ {
		fmt.Fprintf(&b, "var v%d = %d\n", i, i)
	}
	b.WriteString("\nvar events []string\n")
	b.WriteString(`func emit(s string) string { events = append(events, s); return s }` + "\n\n")
	for i := 0; i < 75; i++ {
		fmt.Fprintf(&b, `
func f%d(x int) func() int {
	local := x + v%d
	return func() int {
		return local + v%d + v%d + len(events)
	}
}
`, i, i, (i+1)%150, (i+7)%150)
	}
	b.WriteString("func main() {\n")
	for i := 0; i < 75; i++ {
		fmt.Fprintf(&b, "\t_ = f%d(%d)()\n", i, i)
	}
	b.WriteString(`_ = emit("done")` + "\n")
	b.WriteString("}\n")
	return b.String()
}

// BenchmarkPreprocessCoda measures the full Preprocess call (preprocess1 +
// the four coda passes) on representative sources. Since preprocess1
// dominates total cost, expect only a modest end-to-end speedup from the
// coda-merge refactor — the gains are more visible in per-op allocations
// and in CPU profiles.
func BenchmarkPreprocessCoda(b *testing.B) {
	names := []string{"small", "medium", "large"}
	for _, name := range names {
		src := benchCodaSources[name]
		fname := name + ".gno"
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Re-parse per iteration: Preprocess mutates the node,
				// and FileNode.Copy() loses FileName/StaticBlock. Parse
				// time is excluded from the timer via StopTimer/StartTimer.
				b.StopTimer()
				m := NewMachine("bench", nil)
				fn := m.MustParseFile(fname, src)
				fset := &FileSet{Files: []*FileNode{fn}}
				pn := NewPackageNode("bench", "gno.land/p/bench", fset)
				m.Store.SetBlockNode(pn)
				b.StartTimer()

				PredefineFileSet(m.Store, pn, fset)
				_ = Preprocess(m.Store, pn, fn).(*FileNode)
			}
		})
	}
}
