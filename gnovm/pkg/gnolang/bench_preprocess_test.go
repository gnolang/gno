package gnolang

import (
	"fmt"
	"testing"

	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// bench_preprocess_test.go: Go benchmark for the Preprocess pass.
// Reports ns/op(pure) (whole pass) and ns/<CodeName> per
// PreprocessOp (calibration data for preprocessGasCosts). Run with:
//
//	go test -run=NONE -bench=BenchmarkPreprocess_Corpus -benchtime=1000x \
//	    ./gnovm/pkg/gnolang/
//
// Per-code numbers include the time.Now() overhead from
// StartPreprocess+StopPreprocess (~94 ns/visit on M1 Pro), so sub-
// 200 ns entries are dominated by instrumentation. Production
// instrumentation is off so the over-charge is safe (DoS-safe). See
// ADR for details.

// preprocessBenchCorpus exercises every reachable preprocess code
// path (chan/select/send/go are parse-rejected, so their codes stay
// at cost 0). Each iteration re-adds it at a unique pkg path.
const preprocessBenchCorpus = `package preprocess

type MyStruct struct {
	A int
	B string
}

type MyIface interface {
	M() int
}

type MyArray [2]int
type MySlice []int
type MyMap map[string]int
type MyFunc func(int) int

var globalVar int
const globalConst = 1

func (s MyStruct) M() int { return s.A }

// sideEffect is called as a standalone statement to exercise
// PreprocessLeaveExprStmt.
func sideEffect() { globalVar++ }

func RunAll() int {
	var x int = 1
	y := 2
	x = y + globalConst

	neg := -x
	ok := !false

	ms := MyStruct{A: x, B: "b"}
	m := map[string]int{"a": x}
	s := []int{1, 2, 3}
	a := [2]int{1, 2}

	_ = m["a"]
	_ = s[1:2]
	_ = a[1]

	_ = ms.A

	px := &x
	_ = *px

	f := func(n int) int { return n + 1 }
	_ = f(3)

	var iv interface{} = x
	_ = iv.(int)

	if z := x + y; z > 0 {
		x = z
	} else if z == 0 {
		x = 0
	} else {
		x = -1
	}

	for i := 0; i < 2; i++ {
		if i == 1 {
			continue
		}
		x += i
	}

	for i, v := range s {
		x += i + v
	}

	switch x {
	case 1:
		x++
		fallthrough
	case 2:
		x++
	default:
		x++
	}

	switch v := iv.(type) {
	case int:
		x += v
	default:
		x += 0
	}

Loop:
	for i := 0; i < 3; i++ {
		if i == 2 {
			break Loop
		}
	}

	// Explicit block statement — exercises BlockStmt (enter/leave).
	{
		inner := x * 2
		x += inner
	}

	// Defer statement — exercises PreprocessLeaveDeferStmt.
	defer func() {
		_ = x
	}()

	// Empty statement — exercises PreprocessLeaveEmptyStmt.
	;

	// Expression statement — exercises PreprocessLeaveExprStmt.
	sideEffect()

	// Declaration statement (local type) — PreprocessLeaveDeclStmt.
	type localAlias = int
	var _ localAlias = x

	if x > 0 {
		goto done
	}
	x = 0

done:
	globalVar = x
	return x + neg
}
`

// BenchmarkPreprocess_Corpus measures the cost of a full Preprocess
// pass plus per-PreprocessOp averages. Fresh machine + mempackage
// per iteration so store growth doesn't skew measurements.
func BenchmarkPreprocess_Corpus(b *testing.B) {
	SetPreprocessTiming(true)
	defer SetPreprocessTiming(false)

	bm.ResetPreprocess()

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for i := 0; i < b.N; i++ {
		pkgPath := fmt.Sprintf("gno.land/r/x/benchmark/preprocess/p%d", i)
		mpkg := &std.MemPackage{
			Name: "preprocess",
			Path: pkgPath,
			Type: MPUserAll,
			Files: []*std.MemFile{
				{Name: "gnomod.toml", Body: GenGnoModLatest(pkgPath)},
				{Name: "preprocess.gno", Body: preprocessBenchCorpus},
			},
		}
		m := benchMachine()

		bm.SwitchOpCode(bmTarget)
		_, _ = m.RunMemPackage(mpkg, false)
		bm.SwitchOpCode(bmSetup)

		m.Release()
	}

	// Report standard metrics (ns/op(pure), alloc-gas/op).
	reportBenchops(b)

	// Plus per-PreprocessOp averages for copying into
	// preprocessGasCosts in preprocess.go.
	for code := 1; code < 256; code++ {
		count := bm.PreprocessCount(bm.PreprocessOp(code))
		if count == 0 {
			continue
		}
		avgNs := float64(bm.PreprocessAccumDur(bm.PreprocessOp(code)).Nanoseconds()) / float64(count)
		b.ReportMetric(avgNs, "ns/"+bm.PreprocessCodeString(bm.PreprocessOp(code)))
	}
}
