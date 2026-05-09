package gnolang

import (
	"fmt"
	"strings"
	"testing"

	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// bench_preprocess_test.go: Per-PreprocessOp microbenchmarks for the
// Preprocess pass.
//
// One BenchmarkPreprocess<Code> per PreprocessOp, parameterized by visit
// count via b.Run("N=…"). Each iteration constructs minimal Gno source
// containing N copies of the construct that triggers the target visit,
// then times PredefineFileSet + Preprocess wrapped in bm.SwitchOpCode
// markers — same pattern as bench_ops_test.go.
//
// Per-visit cost is extracted offline by gen_preprocess_table.py from
// the slope of `ns/op(pure)` over N.
//
// Run all:
//
//	make -C gnovm calibrate.preprocess SCALE=<host_factor>
//
// Run one:
//
//	go test -run=NONE -bench=BenchmarkPreprocessLeaveStructTypeExpr \
//	    -benchtime=200x -count=3 ./gnovm/pkg/gnolang/

// benchPreprocess runs b.N iterations of PredefineFileSet + Preprocess
// on `src`, with bmTarget timing wrapping the two preprocess phases.
// Setup work (parse, machine ctor, mempackage construction) runs in
// bmSetup and is excluded from ns/op(pure).
func benchPreprocess(b *testing.B, src string) {
	b.Helper()

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for i := 0; i < b.N; i++ {
		pkgPath := fmt.Sprintf("gno.land/r/x/bench/preprocess/p%d", i)
		mpkg := &std.MemPackage{
			Name: "bench",
			Path: pkgPath,
			Type: MPUserAll,
			Files: []*std.MemFile{
				{Name: "gnomod.toml", Body: GenGnoModLatest(pkgPath)},
				{Name: "bench.gno", Body: src},
			},
		}
		m := benchMachine()
		fset := m.ParseMemPackage(mpkg)

		bm.SwitchOpCode(bmTarget)
		m.PreprocessFiles(mpkg.Name, mpkg.Path, fset, false, false, "")
		bm.SwitchOpCode(bmSetup)

		m.Release()
	}
	reportBenchops(b)
}

// benchPreprocessParam runs the bench at multiple N values via b.Run.
// gen_preprocess_table.py slope-fits across these points to extract the
// per-visit cost.
func benchPreprocessParam(b *testing.B, build func(n int) string) {
	for _, n := range []int{1, 8, 64} {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			benchPreprocess(b, build(n))
		})
	}
}

// repeat returns block of N "stmt\n" lines, indented for placement
// inside a function body or file-level decl block.
func repeat(stmt string, n, indent int) string {
	var sb strings.Builder
	pad := strings.Repeat("\t", indent)
	for i := 0; i < n; i++ {
		sb.WriteString(pad)
		// Substitute %d placeholder so callers can produce unique names.
		sb.WriteString(fmt.Sprintf(stmt, i))
		sb.WriteString("\n")
	}
	return sb.String()
}

// wrapFunc places body inside `func bench<i>() { body }` so that
// statements can be top-level-friendly. The %s for i is the function
// counter so repeated calls don't redeclare. Use n (declarations
// per source) to derive the function name.
func wrapFunc(body string) string {
	return "package bench\n\nfunc bench() {\n" + body + "}\n"
}

// wrapFile wraps top-level decls (no surrounding func).
func wrapFile(decls string) string {
	return "package bench\n\n" + decls
}

// ---------------------------------------------------------------------------
// Type expressions (TRANS_LEAVE)
// Strategy: N parallel `type T<i> = …` aliases at file level. The target
// type-expr code fires once per alias.
// ---------------------------------------------------------------------------

func BenchmarkPreprocessLeaveStructTypeExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFile(repeat("type T%d = struct{ a int; b string }", n, 0))
	})
}

func BenchmarkPreprocessLeaveArrayTypeExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFile(repeat("type T%d = [4]int", n, 0))
	})
}

func BenchmarkPreprocessLeaveSliceTypeExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFile(repeat("type T%d = []int", n, 0))
	})
}

func BenchmarkPreprocessLeaveMapTypeExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFile(repeat("type T%d = map[string]int", n, 0))
	})
}

func BenchmarkPreprocessLeaveInterfaceTypeExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFile(repeat("type T%d interface{ M() int }", n, 0))
	})
}

func BenchmarkPreprocessLeaveFuncTypeExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFile(repeat("type T%d = func(int) int", n, 0))
	})
}

func BenchmarkPreprocessLeaveFieldTypeExpr(b *testing.B) {
	// FieldTypeExpr fires once per field. Vary number of fields in one
	// struct rather than number of structs, so other type-expr codes
	// stay constant.
	benchPreprocessParam(b, func(n int) string {
		fields := ""
		for i := 0; i < n; i++ {
			fields += fmt.Sprintf("\tf%d int\n", i)
		}
		return wrapFile("type T = struct{\n" + fields + "}")
	})
}

// ---------------------------------------------------------------------------
// Top-level declarations
// ---------------------------------------------------------------------------

func BenchmarkPreprocessLeaveTypeDecl(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFile(repeat("type T%d = int", n, 0))
	})
}

func BenchmarkPreprocessEnterTypeDecl(b *testing.B) {
	// Same source as LeaveTypeDecl — slope captures the sum of
	// EnterTypeDecl + LeaveTypeDecl + LeaveNameExpr (for `int`).
	benchPreprocessParam(b, func(n int) string {
		return wrapFile(repeat("type T%d = int", n, 0))
	})
}

func BenchmarkPreprocessLeaveValueDecl(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFile(repeat("var V%d = 1", n, 0))
	})
}

func BenchmarkPreprocessEnterValueDecl(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFile(repeat("var V%d = 1", n, 0))
	})
}

func BenchmarkPreprocessLeaveFuncDecl(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFile(repeat("func F%d() {}", n, 0))
	})
}

func BenchmarkPreprocessEnterFuncDecl(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFile(repeat("func F%d() {}", n, 0))
	})
}

func BenchmarkPreprocessBlockFuncDecl(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFile(repeat("func F%d() {}", n, 0))
	})
}

// ---------------------------------------------------------------------------
// Function-body statements (TRANS_LEAVE / TRANS_BLOCK)
// All wrapped in `func bench()` so wrapping AST is constant.
// ---------------------------------------------------------------------------

func BenchmarkPreprocessLeaveAssignStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\tx := 0\n" + repeat("x = %d", n, 1) + "\t_ = x\n")
	})
}

func BenchmarkPreprocessLeaveIncDecStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\tx := 0\n" + repeat("x++ // %d", n, 1) + "\t_ = x\n")
	})
}

func BenchmarkPreprocessLeaveBranchStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\tfor {\n" + repeat("\t\tif true { break } // %d", n, 0) + "\t}\n")
	})
}

func BenchmarkPreprocessLeaveForStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("for i := 0; i < 1; i++ { _ = i + %d }", n, 1))
	})
}

func BenchmarkPreprocessBlockForStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("for i := 0; i < 1; i++ { _ = i + %d }", n, 1))
	})
}

func BenchmarkPreprocessLeaveIfStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("if true { _ = %d }", n, 1))
	})
}

func BenchmarkPreprocessBlockIfStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("if true { _ = %d }", n, 1))
	})
}

func BenchmarkPreprocessLeaveIfCaseStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("if true { _ = %d } else { _ = 0 }", n, 1))
	})
}

func BenchmarkPreprocessBlockIfCaseStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("if true { _ = %d } else { _ = 0 }", n, 1))
	})
}

func BenchmarkPreprocessLeaveRangeStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\ts := []int{1}\n" + repeat("for _, v := range s { _ = v + %d }", n, 1))
	})
}

func BenchmarkPreprocessBlockRangeStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\ts := []int{1}\n" + repeat("for _, v := range s { _ = v + %d }", n, 1))
	})
}

func BenchmarkPreprocessLeaveSwitchStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\tx := 0\n" + repeat("switch x { case %d: }", n, 1) + "\t_ = x\n")
	})
}

func BenchmarkPreprocessBlockSwitchStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\tx := 0\n" + repeat("switch x { case %d: }", n, 1) + "\t_ = x\n")
	})
}

func BenchmarkPreprocessBlock2SwitchStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\tx := 0\n" + repeat("switch x { case %d: }", n, 1) + "\t_ = x\n")
	})
}

func BenchmarkPreprocessLeaveSwitchClauseStmt(b *testing.B) {
	// One switch with N cases; LeaveSwitchClauseStmt fires per case.
	benchPreprocessParam(b, func(n int) string {
		var cases strings.Builder
		for i := 0; i < n; i++ {
			fmt.Fprintf(&cases, "\tcase %d:\n", i)
		}
		return wrapFunc("\tx := 0\n\tswitch x {\n" + cases.String() + "\t}\n\t_ = x\n")
	})
}

func BenchmarkPreprocessBlockSwitchClauseStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		var cases strings.Builder
		for i := 0; i < n; i++ {
			fmt.Fprintf(&cases, "\tcase %d:\n", i)
		}
		return wrapFunc("\tx := 0\n\tswitch x {\n" + cases.String() + "\t}\n\t_ = x\n")
	})
}

func BenchmarkPreprocessLeaveReturnStmt(b *testing.B) {
	// One func per return (since you can't have multiple unconditional
	// returns in one block).
	benchPreprocessParam(b, func(n int) string {
		var fns strings.Builder
		for i := 0; i < n; i++ {
			fmt.Fprintf(&fns, "func R%d() int { return %d }\n", i, i)
		}
		return wrapFile(fns.String())
	})
}

func BenchmarkPreprocessLeaveDeferStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("defer func(){ _ = %d }()", n, 1))
	})
}

func BenchmarkPreprocessLeaveExprStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\tg := func(int) {}\n" + repeat("g(%d)", n, 1))
	})
}

func BenchmarkPreprocessLeaveDeclStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("var v%d = 0; _ = v%[1]d", n, 1))
	})
}

func BenchmarkPreprocessLeaveBlockStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("{ _ = %d }", n, 1))
	})
}

func BenchmarkPreprocessBlockBlockStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("{ _ = %d }", n, 1))
	})
}

func BenchmarkPreprocessLeaveEmptyStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		// An empty statement can appear after a label.
		var sb strings.Builder
		for i := 0; i < n; i++ {
			fmt.Fprintf(&sb, "\tL%d: ;\n\t_ = 0 // %d\n", i, i)
		}
		return wrapFunc(sb.String())
	})
}

// ---------------------------------------------------------------------------
// Expressions
// ---------------------------------------------------------------------------

func BenchmarkPreprocessLeaveBinaryExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("_ = 1 + %d", n, 1))
	})
}

func BenchmarkPreprocessLeaveUnaryExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("_ = -%d", n, 1))
	})
}

func BenchmarkPreprocessLeaveBasicLitExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("_ = %d", n, 1))
	})
}

func BenchmarkPreprocessLeaveNameExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\ta := 1\n" + repeat("_ = a // %d", n, 1))
	})
}

func BenchmarkPreprocessLeaveCallExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\tg := func() int { return 1 }\n" + repeat("_ = g() + %d", n, 1))
	})
}

func BenchmarkPreprocessLeaveIndexExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\ta := [1]int{0}\n" + repeat("_ = a[0] + %d", n, 1))
	})
}

func BenchmarkPreprocessLeaveSliceExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\ts := []int{0}\n" + repeat("_ = s[:] // %d", n, 1))
	})
}

func BenchmarkPreprocessLeaveTypeAssertExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\tvar i interface{} = 0\n" + repeat("_, _ = i.(int) // %d", n, 1))
	})
}

func BenchmarkPreprocessLeaveCompositeLitExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("_ = []int{%d}", n, 1))
	})
}

func BenchmarkPreprocessLeaveStarExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\tv := 0; p := &v\n" + repeat("_ = *p + %d", n, 1))
	})
}

func BenchmarkPreprocessLeaveSelectorExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\ts := struct{ a int }{a: 0}\n" + repeat("_ = s.a + %d", n, 1))
	})
}

func BenchmarkPreprocessLeaveFuncLitExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("_ = func() int { return %d }", n, 1))
	})
}

func BenchmarkPreprocessBlockFuncLitExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("_ = func() int { return %d }", n, 1))
	})
}

// ---------------------------------------------------------------------------
// EnterAssignStmt: AssignStmt-specific TRANS_ENTER work in preprocess1.
// ---------------------------------------------------------------------------

func BenchmarkPreprocessEnterAssignStmt(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\tx := 0\n" + repeat("x = %d", n, 1) + "\t_ = x\n")
	})
}

func BenchmarkPreprocessEnterFuncTypeExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFile(repeat("type T%d = func(int) int", n, 0))
	})
}

// ---------------------------------------------------------------------------
// Codes that are flat per file (N=1 only) or hard to deterministically
// trigger N times. Run a single benchmark and let gen_preprocess_table.py
// emit them as flat-with-warning entries.
// ---------------------------------------------------------------------------

func BenchmarkPreprocessLeaveFileNode(b *testing.B) {
	benchPreprocess(b, wrapFile("var X = 1\n"))
}

func BenchmarkPreprocessBlockFileNode(b *testing.B) {
	benchPreprocess(b, wrapFile("var X = 1\n"))
}

func BenchmarkPreprocessLeaveImportDecl(b *testing.B) {
	// Imports in the bench machine require a stub package; just measure
	// the simplest legal source. Approximated downstream.
	benchPreprocess(b, wrapFile("var X = 1\n"))
}

func BenchmarkPreprocessEnterImportDecl(b *testing.B) {
	benchPreprocess(b, wrapFile("var X = 1\n"))
}

// LeaveConstExpr fires when a constant-folded sub-tree is re-preprocessed.
// Hard to deterministically trigger; fall back to a representative source
// that exercises constant folding in a binary expression.
func BenchmarkPreprocessLeaveConstExpr(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("const c%d = 1 + 1", n, 1))
	})
}

func BenchmarkPreprocessLeaveRefExpr(b *testing.B) {
	// RefExpr is internal-use; use & operator to force a Ref-style visit.
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc("\tv := 0\n" + repeat("_ = &v // %d", n, 1))
	})
}

// ---------------------------------------------------------------------------
// Generic stage codes (catch-alls for nodes without per-type work).
// These fire in addition to the per-type codes above; their cost is the
// minimal Transcribe-callback overhead. Source is intentionally simple so
// the slope reflects only generic visits.
// ---------------------------------------------------------------------------

func BenchmarkPreprocessEnter(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("_ = %d", n, 1))
	})
}

func BenchmarkPreprocessBlock(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("_ = %d", n, 1))
	})
}

func BenchmarkPreprocessBlock2(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("_ = %d", n, 1))
	})
}

func BenchmarkPreprocessLeave(b *testing.B) {
	benchPreprocessParam(b, func(n int) string {
		return wrapFunc(repeat("_ = %d", n, 1))
	})
}
