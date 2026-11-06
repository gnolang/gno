package gnolang

import (
	"fmt"
	"io"
	"math"
	"testing"
	"time"

	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

var (
	benchmarkFindFirstClosureFLE     *FuncLitExpr
	benchmarkFindFirstClosureDepth   int
	benchmarkFindFirstClosureFound   bool
	benchmarkFindFirstClosureScanned int
)

func BenchmarkFindFirstClosure(b *testing.B) {
	for _, depth := range []int{1, 8, 32, 128} {
		b.Run(fmt.Sprintf("depth=%d", depth), func(b *testing.B) {
			stop := &FuncDecl{}
			closures := make([]BlockNode, 1, depth+1)
			closures[0] = stop
			for range depth {
				closures = append(closures, &FuncLitExpr{})
			}

			b.Run("closure-only/found", func(b *testing.B) {
				benchmarkFindFirstClosure(b, closures, stop, len(closures))
			})

			mixed := make([]BlockNode, 1, 3*depth+1)
			mixed[0] = stop
			for range depth {
				mixed = append(mixed, &BlockStmt{}, &IfCaseStmt{}, &FuncLitExpr{})
			}
			b.Run("mixed/found", func(b *testing.B) {
				benchmarkFindFirstClosure(b, mixed, stop, len(mixed))
			})

			notFound := make([]BlockNode, 0, 2*depth)
			for range depth {
				notFound = append(notFound, &BlockStmt{}, &IfCaseStmt{})
			}
			b.Run("non-closure-faux/not-found", func(b *testing.B) {
				benchmarkFindFirstClosure(b, notFound, stop, len(notFound))
			})

			b.Run("closure-only/cumulative", func(b *testing.B) {
				frames := (depth + 1) * (depth + 2) / 2
				b.ResetTimer()
				for range b.N {
					dbn := BlockNode(stop)
					for {
						flx, gotDepth, found, scanned := findFirstClosure(closures, dbn)
						benchmarkFindFirstClosureFLE = flx
						benchmarkFindFirstClosureDepth = gotDepth
						benchmarkFindFirstClosureFound = found
						benchmarkFindFirstClosureScanned = scanned
						if !found {
							break
						}
						dbn = flx
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*frames), "ns/frame")
			})
		})
	}
}

// BenchmarkClosureCapture preprocesses captureNamesSrc with one name read
// names times and with names distinct names, and reports the time difference
// per extra capture. Both sources have the same bytes and the same scans, so
// the difference is what addHeapCapture spends on each new capture, garbage
// collection included.
func BenchmarkClosureCapture(b *testing.B) {
	for _, c := range []struct{ names, depth int }{{32, 8}, {200, 20}, {2000, 100}} {
		b.Run(fmt.Sprintf("names=%d/depth=%d", c.names, c.depth), func(b *testing.B) {
			same := captureNamesSrc(c.names, c.depth, false)
			distinct := captureNamesSrc(c.names, c.depth, true)
			var spent time.Duration
			for range b.N {
				spent += timePreprocess(b, distinct) - timePreprocess(b, same)
			}
			b.ReportMetric(float64(spent.Nanoseconds())/float64((c.names-1)*c.depth*b.N), "ns/capture")
		})
	}
}

// timePreprocess returns the wall time PreprocessFiles takes on src, with
// the parse and the store setup left out.
func timePreprocess(b *testing.B, src string) time.Duration {
	b.Helper()
	b.StopTimer()
	defer b.StartTimer()
	st, _ := newPreprocessAllocTestStore(b, math.MaxInt64, stypes.NewInfiniteGasMeter())
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "gno.land/r/test/closure_scan",
		Store:   st,
		Output:  io.Discard,
		Alloc:   NewAllocator(math.MaxInt64),
	})
	defer m.Release()
	fset := &FileSet{Files: []*FileNode{m.MustParseFile("closure_scan.gno", src)}}
	b.StartTimer()
	start := time.Now()
	m.PreprocessFiles("closure_scan", "gno.land/r/test/closure_scan", fset, false, false)
	spent := time.Since(start)
	b.StopTimer()
	return spent
}

func benchmarkFindFirstClosure(b *testing.B, stack []BlockNode, stop BlockNode, frames int) {
	b.Helper()
	b.ResetTimer()
	for range b.N {
		benchmarkFindFirstClosureFLE, benchmarkFindFirstClosureDepth, benchmarkFindFirstClosureFound, benchmarkFindFirstClosureScanned = findFirstClosure(stack, stop)
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*frames), "ns/frame")
}
