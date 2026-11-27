package gnolang

import (
	"fmt"
	"io"
	"math"
	"strings"
	"testing"

	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

type closureScanGasMeter struct {
	stypes.GasMeter
	cpu stypes.Gas
}

func (m *closureScanGasMeter) ConsumeGas(amount stypes.Gas, descriptor string) {
	m.GasMeter.ConsumeGas(amount, descriptor)
	if descriptor == "CPUCycles" {
		m.cpu += amount
	}
}

func TestPreprocessClosureScanGas(t *testing.T) {
	t.Run("superlinear_in_depth", func(t *testing.T) {
		g8 := measurePreprocessClosureScanGas(t, 8, 1)
		g16 := measurePreprocessClosureScanGas(t, 16, 1)
		g32 := measurePreprocessClosureScanGas(t, 32, 1)
		t.Logf("closure-scan CPU gas: depth=8 %d, depth=16 %d, depth=32 %d", g8, g16, g32)

		require.Greater(t, g8, int64(0))
		require.Greater(t, g32-g16, 2*(g16-g8))
	})

	t.Run("linear_in_references", func(t *testing.T) {
		one := measurePreprocessClosureScanGas(t, 16, 1)
		four := measurePreprocessClosureScanGas(t, 16, 4)
		t.Logf("closure-scan CPU gas: refs=1 %d, refs=4 %d", one, four)

		require.Greater(t, one, int64(0))
		require.Equal(t, 4*one, four)
	})

	// A scan that finds no closure walks the whole stack too, so it must be
	// charged like one that finds a closure: references nested in plain
	// blocks cost more the deeper they sit.
	t.Run("charges_scans_that_find_no_closure", func(t *testing.T) {
		added := func(blocks, refs int) int64 {
			return closureScanGas(t, nestedBlockScanSrc(blocks, refs)) - closureScanGas(t, nestedBlockScanSrc(blocks, 0))
		}
		one := added(32, 1) - added(0, 1)
		four := added(32, 4) - added(0, 4)
		t.Logf("closure-scan CPU gas 32 blocks deep without closures: refs=1 %d, refs=4 %d", one, four)

		require.Greater(t, one, int64(0))
		require.Equal(t, 4*one, four)
	})

	// append and copy preprocess a string argument a second time to convert
	// it to []byte. Closures inside it must pay on that pass like anywhere else.
	t.Run("append_and_copy_pay_like_plain", func(t *testing.T) {
		added := func(shape string) int64 {
			return closureScanGas(t, builtinArgClosureScanSrc(shape, 32, 64)) -
				closureScanGas(t, builtinArgClosureScanSrc(shape, 32, 0))
		}
		plain := added("plain")
		t.Logf("closure-scan CPU gas added by 64 references at depth 32: %d", plain)

		require.Equal(t, plain, added("append"))
		require.Equal(t, plain, added("copy"))
	})

	// A closure captures each name once, so distinct names pay a capture
	// apiece under every closure, while repeated reads of one name share it.
	// Each closure's block also pays for the slots it copies as it grows,
	// which past blockValuesGrowThreshold happens every
	// blockValuesGrowIncrement captures.
	t.Run("charges_each_new_capture", func(t *testing.T) {
		for _, c := range []struct{ names, depth int }{{32, 8}, {1200, 2}} {
			same := closureScanGas(t, captureNamesSrc(c.names, c.depth, false))
			distinct := closureScanGas(t, captureNamesSrc(c.names, c.depth, true))
			t.Logf("CPU gas for %d names under %d closures: one name %d, distinct %d", c.names, c.depth, same, distinct)

			captures := int64((c.names - 1) * c.depth)
			copies := int64(c.depth * growthCopies(c.names))
			require.Equal(t, captures*OpCPUSlopePreprocessClosureCapture+copies*OpCPUSlopePreprocessCaptureGrowCopy, distinct-same)
		}
	})
}

// growthCopies counts the slots growBlockValues copies while an empty block
// takes n defines one at a time, as a closure's block does for n captures.
func growthCopies(n int) (copied int) {
	var values []TypedValue
	for i := 1; i <= n; i++ {
		if cap(values) < i {
			copied += len(values)
		}
		values = growBlockValues(values, i)
	}
	return copied
}

// measurePreprocessClosureScanGas returns the scan gas of refs reads of x
// under depth closures. Any read of x adds the same depth captures, one per
// closure, so their charge is taken out.
func measurePreprocessClosureScanGas(t *testing.T, depth, refs int) int64 {
	t.Helper()
	if refs == 0 {
		return 0
	}
	perRefWithoutClosures := rawPreprocessClosureScanGas(t, 0, refs) - rawPreprocessClosureScanGas(t, 0, 0)
	captures := int64(depth) * OpCPUSlopePreprocessClosureCapture
	return rawPreprocessClosureScanGas(t, depth, refs) - rawPreprocessClosureScanGas(t, depth, 0) - perRefWithoutClosures - captures
}

// captureNamesSrc reads names parameters from inside depth nested closures,
// each once when distinct, or the first one names times otherwise. Both
// shapes have the same bytes and the same scans.
func captureNamesSrc(names, depth int, distinct bool) string {
	var params, refs strings.Builder
	for i := range names {
		if i > 0 {
			params.WriteString(", ")
		}
		fmt.Fprintf(&params, "a%05d int", i)
		ref := 0
		if distinct {
			ref = i
		}
		fmt.Fprintf(&refs, "_ = a%05d\n", ref)
	}
	return "package closure_scan\n\nfunc measure(" + params.String() + ") {\n" +
		strings.Repeat("_ = func() {\n", depth) + refs.String() + strings.Repeat("}\n", depth) + "}\n"
}

func rawPreprocessClosureScanGas(t *testing.T, depth, refs int) int64 {
	t.Helper()

	var src strings.Builder
	src.WriteString("package closure_scan\n\nfunc measure(x int) {\n")
	for i := range depth {
		fmt.Fprintf(&src, "%s_ = func() {\n", strings.Repeat("\t", i+1))
	}
	for range refs {
		fmt.Fprintf(&src, "%s_ = x\n", strings.Repeat("\t", depth+1))
	}
	for i := depth; i > 0; i-- {
		fmt.Fprintf(&src, "%s}\n", strings.Repeat("\t", i))
	}
	src.WriteString("}\n")
	return closureScanGas(t, src.String())
}

// nestedBlockScanSrc references the parameter x refs times from inside
// blocks nested plain blocks, with no closure around them.
func nestedBlockScanSrc(blocks, refs int) string {
	return "package closure_scan\n\nfunc measure(x int) {\n" + strings.Repeat("{\n", blocks) +
		strings.Repeat("_ = x\n", refs) + strings.Repeat("}\n", blocks) + "}\n"
}

// builtinArgClosureScanSrc nests the closures of rawPreprocessClosureScanGas
// in a func literal returning a string, and passes its result as a plain
// statement, as the string argument of append, or as that of copy.
func builtinArgClosureScanSrc(shape string, depth, refs int) string {
	arg := "func() string {\n" + strings.Repeat("_ = func() {\n", depth) +
		strings.Repeat("_ = x\n", refs) + strings.Repeat("}\n", depth) + "return \"\"\n}()"
	body := map[string]string{
		"plain":  "_ = %s\n",
		"append": "var b []byte\nb = append(b, %s...)\n_ = b\n",
		"copy":   "b := make([]byte, 1)\ncopy(b, %s)\n",
	}[shape]
	return "package closure_scan\n\nfunc measure(x int) {\n" + fmt.Sprintf(body, arg) + "}\n"
}

// closureScanGas preprocesses src and returns the CPU gas it charged.
func closureScanGas(t *testing.T, src string) int64 {
	t.Helper()

	meter := &closureScanGasMeter{GasMeter: stypes.NewInfiniteGasMeter()}
	st, _ := newPreprocessAllocTestStore(t, 64*1024*1024, meter)
	t.Cleanup(func() { st.SetPreprocessAllocator(nil) })
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "gno.land/r/test/closure_scan",
		Store:   st,
		Output:  io.Discard,
		Alloc:   NewAllocator(math.MaxInt64),
	})
	t.Cleanup(m.Release)

	file := m.MustParseFile("closure_scan.gno", src)
	m.PreprocessFiles("closure_scan", "gno.land/r/test/closure_scan", &FileSet{Files: []*FileNode{file}}, false, false)
	return meter.cpu
}

// TestPreprocessClosureScanGasFauxFrames checks that the faux child blocks of
// if and switch statements are charged like any other frame.
func TestPreprocessClosureScanGasFauxFrames(t *testing.T) {
	added := func(open string, blocks int) int64 {
		src := func(refs int) string {
			return "package closure_scan\n\nfunc measure(x int) {\n" + strings.Repeat(open+"\n", blocks) +
				strings.Repeat("_ = x\n", refs) + strings.Repeat("}\n", blocks) + "}\n"
		}
		return closureScanGas(t, src(1)) - closureScanGas(t, src(0))
	}
	plain := added("{", 32) - added("{", 0)
	ifs := added("if true {", 32) - added("if true {", 0)
	t.Logf("closure-scan CPU gas 32 levels deep: plain blocks %d, if blocks %d", plain, ifs)
	require.Equal(t, 2*plain, ifs) // an if is two frames: IfStmt and IfCaseStmt.
}
