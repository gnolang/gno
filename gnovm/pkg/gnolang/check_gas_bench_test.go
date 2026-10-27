package gnolang_test

import (
	"io"
	"math"
	"path/filepath"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// checkGasCorpus is a fixed set of larger example packages, loaded and
// preprocessed once (outside the timed region) through the test store.
var checkGasCorpus = []string{
	"gno.land/r/gnoland/boards2/v0",
	"gno.land/r/nt/commondao/v0",
	"gno.land/p/onbloc/json/v0",
	"gno.land/p/onbloc/uint256/v0",
	"gno.land/p/nt/markdown/sanitize/v0",
	"gno.land/p/nt/commondao/v0",
	"gno.land/r/gov/dao/impl/v0",
	"gno.land/p/jeronimoalbi/expect/v0",
	"gno.land/r/sys/params",
	"gno.land/p/gnoland/boards/v0",
	"gno.land/r/gnops/valopers",
	"gno.land/p/nt/bptree/v0",
	"gno.land/p/nt/avl/v0",
	"gno.land/p/nt/ufmt/v0",
}

type checkGasFamily struct {
	name  string
	slope int64 // gas per charged node; nodes = gas / slope
	count int64 // direct node count when the walk charges other gas too
	decls int64 // declarations an unmetered per-package setup runs over
	run   func(gm storetypes.GasMeter)
}

// BenchmarkCheckGas times each metered checker walk over the corpus and
// reports ns per charged node (OpCPUSlopeCheckNode, or
// OpCPUSlopeInitCycleVisit for initCycles). Convert to gas with the
// reference-machine factor (machine.go: 1 gas = 1 ns on the Xeon 8168;
// the M-series dev box measures ~2.1x faster on the flat ops).
//
//	go test -run XXX -bench BenchmarkCheckGas -benchmem -count 3 ./gnovm/pkg/gnolang/
func BenchmarkCheckGas(b *testing.B) {
	rootDir, err := filepath.Abs("../../../")
	if err != nil {
		b.Fatal(err)
	}
	_, store := test.StoreWithOptions(rootDir, io.Discard, test.StoreOptions{WithExamples: true})
	var pkgs []*gno.PackageNode
	for _, path := range checkGasCorpus {
		pv := store.GetPackage(path, true)
		if pv == nil {
			b.Fatalf("package %q not loaded", path)
		}
		pkgs = append(pkgs, pv.GetPackageNode(store))
	}
	alloc := gno.NewAllocator(math.MaxInt64)
	store.SetPreprocessAllocator(alloc)
	pkgName := func(pkgPath string) gno.Name {
		if pv := store.GetPackage(pkgPath, false); pv != nil {
			return pv.PkgName
		}
		return ""
	}

	// Collected once: the functions, type declarations and switch
	// statements the walks run over.
	type funcOf struct {
		file *gno.FileNode
		fd   *gno.FuncDecl
	}
	var funcs []funcOf
	type typeOf struct {
		file *gno.FileNode
		dt   *gno.DeclaredType
		src  gno.Expr
	}
	var types []typeOf
	type nameOf struct {
		file *gno.FileNode
		nx   *gno.NameExpr
	}
	var names []nameOf
	var switches []*gno.SwitchStmt
	var walkNodes, cases, decls int64
	for _, pn := range pkgs {
		for _, fn := range pn.FileSet.Files {
			for _, d := range fn.Decls {
				decls++
				switch d := d.(type) {
				case *gno.FuncDecl:
					funcs = append(funcs, funcOf{fn, d})
					if !d.IsMethod {
						names = append(names, nameOf{fn, &d.NameExpr})
					}
					gno.Transcribe(d, func(_ []gno.Node, _ gno.TransField, _ int, n gno.Node, stage gno.TransStage) (gno.Node, gno.TransCtrl) {
						if stage == gno.TRANS_ENTER {
							walkNodes++
							if ss, ok := n.(*gno.SwitchStmt); ok {
								switches = append(switches, ss)
								for i := range ss.Clauses {
									cases += int64(len(ss.Clauses[i].Cases))
								}
							}
						}
						return n, gno.TRANS_CONTINUE
					})
				case *gno.TypeDecl:
					names = append(names, nameOf{fn, &d.NameExpr})
					if dt, src := gno.DeclaredTypeOfForTest(d); dt != nil {
						types = append(types, typeOf{fn, dt, src})
					}
				case *gno.ValueDecl:
					for i := range d.NameExprs {
						names = append(names, nameOf{fn, &d.NameExprs[i]})
					}
				}
			}
		}
	}
	if len(funcs) < 100 || len(types) < 20 || len(switches) < 20 {
		b.Fatalf("thin corpus: %d funcs, %d types, %d switches", len(funcs), len(types), len(switches))
	}

	families := []checkGasFamily{
		{name: "transcribe", count: walkNodes, run: func(storetypes.GasMeter) {
			for _, f := range funcs {
				gno.Transcribe(f.fd, func(_ []gno.Node, _ gno.TransField, _ int, n gno.Node, _ gno.TransStage) (gno.Node, gno.TransCtrl) {
					return n, gno.TRANS_CONTINUE
				})
			}
		}},
		{name: "unused", slope: gno.OpCPUSlopeCheckNode, run: func(gm storetypes.GasMeter) {
			for _, f := range funcs {
				gno.CheckUnusedForTest(gm, f.file, f.fd)
			}
		}},
		{name: "unusedImports", slope: gno.OpCPUSlopeCheckNode, run: func(gm storetypes.GasMeter) {
			for _, pn := range pkgs {
				gno.CheckUnusedImportsForTest(gm, pn.FileSet, pkgName)
			}
		}},
		{name: "importClash", slope: gno.OpCPUSlopeCheckNode, run: func(storetypes.GasMeter) {
			for _, n := range names {
				gno.AssertNotImportedForTest(store, n.file, n.nx)
			}
		}},
		{name: "typeShape", slope: gno.OpCPUSlopeCheckNode, run: func(gm storetypes.GasMeter) {
			for _, t := range types {
				gno.CheckTypeShapeForTest(gm, t.file, t.dt)
			}
		}},
		{name: "typeExprShape", slope: gno.OpCPUSlopeCheckNode, run: func(gm storetypes.GasMeter) {
			for _, t := range types {
				gno.CheckTypeExprShapeForTest(gm, t.src)
			}
		}},
		{name: "dupCases", count: cases, run: func(gm storetypes.GasMeter) {
			for _, ss := range switches {
				if ss.IsTypeSwitch {
					gno.DuplicateTypeSwitchCasesForTest(gm, ss.Clauses)
				} else {
					gno.DuplicateSwitchCasesForTest(store, ss.Clauses)
				}
			}
		}},
		{name: "initCycles", slope: gno.OpCPUSlopeInitCycleVisit, decls: decls, run: func(gm storetypes.GasMeter) {
			for _, pn := range pkgs {
				gno.CheckInitCyclesForTest(gm, pn, pn.FileSet)
			}
		}},
	}
	for _, f := range families {
		b.Run(f.name, func(b *testing.B) {
			gm := storetypes.NewGasMeter(math.MaxInt64)
			alloc.SetGasMeter(gm)
			b.ResetTimer()
			for range b.N {
				f.run(gm)
			}
			b.StopTimer()
			nodes := float64(f.count)
			if f.slope != 0 {
				nodes = float64(gm.GasConsumed()) / float64(f.slope*gno.GasFactorCPU) / float64(b.N)
			}
			if nodes == 0 {
				b.Fatal("no nodes charged")
			}
			b.ReportMetric(nodes, "nodes/op")
			if f.decls != 0 {
				b.ReportMetric(float64(f.decls), "decls/op")
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/nodes, "ns/node")
		})
	}
}
