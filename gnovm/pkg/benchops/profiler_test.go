package benchops

import (
	"bytes"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfilerLifecycle(t *testing.T) {
	p := New()

	// Should start in Idle state
	require.Equal(t, StateIdle, p.State())

	// Start should transition to Running
	p.Start()
	require.Equal(t, StateRunning, p.State())

	// Do some work (sleep ensures measurable duration for timing verification)
	p.BeginOp(OpAdd)
	time.Sleep(time.Millisecond)
	p.EndOp()

	// Stop should transition back to Idle and return results
	results := p.Stop()
	require.Equal(t, StateIdle, p.State())
	require.NotNil(t, results)
	assert.NotZero(t, results.Duration)

	// Should be able to immediately reuse
	p.Start()
	require.Equal(t, StateRunning, p.State())
	p.Stop()
	require.Equal(t, StateIdle, p.State())
}

func TestProfilerStartPanicsWhenRunning(t *testing.T) {
	p := New()
	p.Start()
	defer p.Stop() // Clean up

	require.PanicsWithValue(t,
		"benchops: profiler is already running (concurrent access or missing Stop)",
		func() { p.Start() })
}

func TestProfilerConcurrentStartPanics(t *testing.T) {
	p := New()

	var wg sync.WaitGroup
	var panicked atomic.Bool

	// Start from main goroutine
	p.Start()

	// Try to start from another goroutine - should panic
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				panicked.Store(true)
				assert.Equal(t,
					"benchops: profiler is already running (concurrent access or missing Stop)",
					r)
			}
		}()
		p.Start() // should panic
	}()

	wg.Wait()
	p.Stop() // Clean up

	assert.True(t, panicked.Load(), "expected concurrent Start to panic")
}

func TestProfilerStopPanicsWhenNotRunning(t *testing.T) {
	p := New()

	require.PanicsWithValue(t,
		"benchops: Stop called on non-running profiler (missing Start)",
		func() { p.Stop() })
}

func TestProfilerResetPanicsWhenRunning(t *testing.T) {
	p := New()
	p.Start()
	defer p.Stop() // Clean up

	require.PanicsWithValue(t,
		"benchops: Reset called on running profiler (use Stop() instead)",
		func() { p.Reset() })
}

func TestOpMeasurement(t *testing.T) {
	p := New()
	p.timingEnabled = true // Enable timing for this test
	p.Start()

	// Measure some ops (sleep ensures measurable duration for timing verification)
	for i := 0; i < 10; i++ {
		p.BeginOp(OpAdd)
		time.Sleep(time.Millisecond)
		p.EndOp()
	}

	results := p.Stop()

	stat, ok := results.OpStats["OpAdd"]
	require.True(t, ok, "expected OpAdd in results")
	assert.Equal(t, int64(10), stat.Count)
	assert.NotZero(t, stat.TotalNs, "expected non-zero total duration")
	assert.NotZero(t, stat.AvgNs, "expected non-zero average duration")
}

// TestNestedStoreCalls verifies that nested store operations correctly pause and
// resume opcode timing. Sleeps ensure measurable duration differences.
func TestNestedStoreCalls(t *testing.T) {
	p := New()
	p.Start()

	// Start an opcode
	p.BeginOp(OpCall)
	time.Sleep(time.Millisecond)

	// Nested store calls should pause opcode timing
	p.BeginStore(StoreGetPackage)
	time.Sleep(time.Millisecond)

	// Second level nesting
	p.BeginStore(StoreGetObject)
	time.Sleep(time.Millisecond)
	p.EndStore(100)

	// Third level nesting
	p.BeginStore(StoreGetPackageRealm)
	time.Sleep(time.Millisecond)
	p.EndStore(50)

	p.EndStore(200)

	// Resume and finish opcode
	time.Sleep(time.Millisecond)
	p.EndOp()

	results := p.Stop()

	// Check opcode was recorded
	opStat, ok := results.OpStats["OpCall"]
	require.True(t, ok, "expected OpCall in results")
	assert.Equal(t, int64(1), opStat.Count)

	// Check all store ops were recorded
	stores := []string{"StoreGetPackage", "StoreGetObject", "StoreGetPackageRealm"}
	for _, name := range stores {
		stat, ok := results.StoreStats[name]
		require.True(t, ok, "expected %s in results", name)
		assert.Equal(t, int64(1), stat.Count, "%s count mismatch", name)
	}

	// Check sizes were recorded
	assert.Equal(t, int64(100), results.StoreStats["StoreGetObject"].TotalSize)
}

func TestPanicRecovery(t *testing.T) {
	p := New()
	p.Start()

	// Start an op and some store calls
	p.BeginOp(OpCall)
	p.BeginStore(StoreGetPackage)
	p.BeginStore(StoreGetObject)

	// Simulate panic recovery
	p.Recovery()

	// Should be able to continue measuring
	p.BeginOp(OpAdd)
	time.Sleep(time.Millisecond)
	p.EndOp()

	results := p.Stop()

	// Only OpAdd should be in results (OpCall was not ended)
	_, hasOpCall := results.OpStats["OpCall"]
	assert.False(t, hasOpCall, "OpCall should not be in results after recovery")

	_, hasOpAdd := results.OpStats["OpAdd"]
	assert.True(t, hasOpAdd, "OpAdd should be in results after recovery")
}

func TestResultsJSON(t *testing.T) {
	p := New()
	p.Start()

	p.BeginOp(OpAdd)
	p.SetOpContext(OpContext{
		File:     "test.gno",
		Line:     10,
		FuncName: "add",
		PkgPath:  "gno.land/r/demo/test",
	})
	p.EndOp()

	results := p.Stop()

	var buf bytes.Buffer
	err := results.WriteJSON(&buf)
	require.NoError(t, err)
	assert.NotZero(t, buf.Len(), "expected non-empty JSON output")

	output := buf.String()
	assert.Contains(t, output, "OpAdd")
	assert.Contains(t, output, "LocationStats")
	assert.Contains(t, output, `"file":"test.gno"`)
	assert.Contains(t, output, `"line":10`)
	assert.Contains(t, output, `"func":"add"`)
	assert.Contains(t, output, `"pkg":"gno.land/r/demo/test"`)
	assert.Contains(t, output, `"count":1`)
	assert.Contains(t, output, `"gas":18`)
}

func TestResultsReport(t *testing.T) {
	p := New()
	p.Start()

	p.BeginOp(OpAdd)
	p.EndOp()
	p.BeginStore(StoreGetObject)
	p.EndStore(42)

	results := p.Stop()

	var buf bytes.Buffer
	err := results.WriteReport(&buf, 10)
	require.NoError(t, err)
	assert.NotZero(t, buf.Len(), "expected non-empty report output")

	output := buf.String()
	assert.Contains(t, output, "OpAdd")
	assert.Contains(t, output, "StoreGetObject")
}

func TestEndOpPanicsWithoutBegin(t *testing.T) {
	p := New()
	p.Start()
	defer p.Stop()

	require.PanicsWithValue(t,
		"benchops: EndOp called without matching BeginOp",
		func() { p.EndOp() })
}

func TestEndStorePanicsWithoutBegin(t *testing.T) {
	p := New()
	p.Start()
	defer p.Stop()

	require.PanicsWithValue(t,
		"benchops: EndStore called without matching BeginStore",
		func() { p.EndStore(0) })
}

func TestEndNativePanicsWithoutBegin(t *testing.T) {
	p := New()
	p.Start()
	defer p.Stop()

	require.PanicsWithValue(t,
		"benchops: EndNative called without matching BeginNative",
		func() { p.EndNative() })
}

func TestOpString(t *testing.T) {
	tests := map[string]struct {
		op   Op
		want string
	}{
		"add":     {op: OpAdd, want: "OpAdd"},
		"call":    {op: OpCall, want: "OpCall"},
		"unknown": {op: Op(0xFE), want: "OpUnknown"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.op.String())
		})
	}
}

func TestStoreOpString(t *testing.T) {
	tests := map[string]struct {
		op   StoreOp
		want string
	}{
		"get_object":  {op: StoreGetObject, want: "StoreGetObject"},
		"set_package": {op: StoreSetPackage, want: "StoreSetPackage"},
		"unknown":     {op: StoreOp(0xFE), want: "StoreOpUnknown"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.op.String())
		})
	}
}

func TestNativeOpString(t *testing.T) {
	tests := map[string]struct {
		op   NativeOp
		want string
	}{
		"print":   {op: NativePrint, want: "NativePrint"},
		"print1":  {op: NativePrint1, want: "NativePrint1"},
		"unknown": {op: NativeOp(0xFE), want: "NativeOpUnknown"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.op.String())
		})
	}
}

func TestGetNativePrintCode(t *testing.T) {
	tests := map[string]struct {
		size int
		want NativeOp
	}{
		"1":     {size: 1, want: NativePrint1},
		"1000":  {size: 1000, want: NativePrint1000},
		"10000": {size: 10000, want: NativePrint1e4},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, GetNativePrintCode(tc.size))
		})
	}
}

func TestGetNativePrintCodePanics(t *testing.T) {
	require.Panics(t, func() { GetNativePrintCode(42) })
}

func TestGetOpGas(t *testing.T) {
	// Test known op
	assert.Equal(t, int64(18), GetOpGas(OpAdd))

	// Test unknown op returns 1
	assert.Equal(t, int64(1), GetOpGas(Op(0xFE)))
}

func TestLocationTracking(t *testing.T) {
	p := New()
	p.Start()

	// Measure ops with different source locations
	for i := 0; i < 5; i++ {
		p.BeginOp(OpAdd)
		p.SetOpContext(OpContext{
			File:     "test.gno",
			Line:     10,
			FuncName: "add",
			PkgPath:  "gno.land/r/demo/test",
		})
		time.Sleep(time.Microsecond)
		p.EndOp()
	}

	for i := 0; i < 3; i++ {
		p.BeginOp(OpMul)
		p.SetOpContext(OpContext{
			File:     "test.gno",
			Line:     20,
			FuncName: "mul",
			PkgPath:  "gno.land/r/demo/test",
		})
		time.Sleep(time.Microsecond)
		p.EndOp()
	}

	results := p.Stop()

	// Check location stats are present
	require.NotNil(t, results.LocationStats)
	require.Len(t, results.LocationStats, 2)

	// Check they're sorted by gas (descending)
	assert.GreaterOrEqual(t, results.LocationStats[0].Gas, results.LocationStats[1].Gas)

	// Check line 10 stats (5 OpAdd = 5 * 18 = 90 gas)
	var line10Stat, line20Stat *LocationStatJSON
	for _, stat := range results.LocationStats {
		if stat.Line == 10 {
			line10Stat = stat
		}
		if stat.Line == 20 {
			line20Stat = stat
		}
	}

	require.NotNil(t, line10Stat)
	assert.Equal(t, "test.gno", line10Stat.File)
	assert.Equal(t, "add", line10Stat.FuncName)
	assert.Equal(t, "gno.land/r/demo/test", line10Stat.PkgPath)
	assert.Equal(t, int64(5), line10Stat.Count)
	assert.Equal(t, int64(5*18), line10Stat.Gas) // 5 OpAdd * 18 gas each

	require.NotNil(t, line20Stat)
	assert.Equal(t, "test.gno", line20Stat.File)
	assert.Equal(t, "mul", line20Stat.FuncName)
	assert.Equal(t, int64(3), line20Stat.Count)
	assert.Equal(t, int64(3*19), line20Stat.Gas) // 3 OpMul * 19 gas each
}

func TestLocationTrackingNoContext(t *testing.T) {
	p := New()
	p.Start()

	// Measure ops without setting context
	p.BeginOp(OpAdd)
	p.EndOp()

	results := p.Stop()

	// No location stats should be recorded
	assert.Nil(t, results.LocationStats)

	// But op stats should be present
	assert.NotNil(t, results.OpStats["OpAdd"])
}

func TestSetOpContextWithoutCurrentOp(t *testing.T) {
	p := New()
	p.Start()

	// Setting context without BeginOp should not panic
	require.NotPanics(t, func() {
		p.SetOpContext(OpContext{
			File: "test.gno",
			Line: 10,
		})
	})

	p.Stop()
}

func TestLocationTrackingReport(t *testing.T) {
	p := New()
	p.Start()

	p.BeginOp(OpAdd)
	p.SetOpContext(OpContext{
		File:     "counter.gno",
		Line:     15,
		FuncName: "Inc",
		PkgPath:  "gno.land/r/demo/counter",
	})
	p.EndOp()

	results := p.Stop()

	var buf bytes.Buffer
	err := results.WriteReport(&buf, 10)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Hot Spots")
	assert.Contains(t, output, "counter.gno:15")
	assert.Contains(t, output, "Inc")
}

func TestWriteGolden(t *testing.T) {
	p := New()
	p.Start()

	// Add ops with context
	p.BeginOp(OpAdd)
	p.SetOpContext(OpContext{
		File:     "test.gno",
		Line:     10,
		FuncName: "add",
		PkgPath:  "gno.land/r/demo/test",
	})
	p.EndOp()

	p.BeginOp(OpMul)
	p.SetOpContext(OpContext{
		File:     "test.gno",
		Line:     20,
		FuncName: "mul",
		PkgPath:  "gno.land/r/demo/test",
	})
	p.EndOp()

	// Add store op
	p.BeginStore(StoreGetObject)
	p.EndStore(42)

	// Add native op
	p.BeginNative(NativePrint)
	p.EndNative()

	results := p.Stop()

	var buf bytes.Buffer
	results.WriteGolden(&buf, 0) // 0 = all sections

	output := buf.String()

	// Should have all deterministic sections
	assert.Contains(t, output, "Opcodes:")
	assert.Contains(t, output, "Store:")
	assert.Contains(t, output, "Native:")
	assert.Contains(t, output, "HotSpots:")

	// Opcodes should be sorted alphabetically
	assert.Contains(t, output, "OpAdd: count=1 gas=18")
	assert.Contains(t, output, "OpMul: count=1 gas=19")

	// Store should be present
	assert.Contains(t, output, "StoreGetObject: count=1 size=42")

	// Native should be present
	assert.Contains(t, output, "NativePrint: count=1")

	// HotSpots should be sorted by file:line
	assert.Contains(t, output, "test.gno:10 add: count=1 gas=18")
	assert.Contains(t, output, "test.gno:20 mul: count=1 gas=19")

	// Verify line 10 comes before line 20 (sorted by file:line)
	pos10 := bytes.Index(buf.Bytes(), []byte("test.gno:10"))
	pos20 := bytes.Index(buf.Bytes(), []byte("test.gno:20"))
	assert.Less(t, pos10, pos20, "HotSpots should be sorted by file:line")
}

func TestWriteGoldenWithFlags(t *testing.T) {
	p := New()
	p.Start()

	p.BeginOp(OpAdd)
	p.EndOp()
	p.BeginStore(StoreGetObject)
	p.EndStore(42)

	results := p.Stop()

	// Test with only Opcodes flag
	var buf bytes.Buffer
	results.WriteGolden(&buf, SectionOpcodes)
	output := buf.String()
	assert.Contains(t, output, "Opcodes:")
	assert.NotContains(t, output, "Store:")
	assert.NotContains(t, output, "Native:")
	assert.NotContains(t, output, "HotSpots:")

	// Test with Opcodes | Store
	buf.Reset()
	results.WriteGolden(&buf, SectionOpcodes|SectionStore)
	output = buf.String()
	assert.Contains(t, output, "Opcodes:")
	assert.Contains(t, output, "Store:")
	assert.NotContains(t, output, "Native:")
	assert.NotContains(t, output, "HotSpots:")
}

func TestSectionFlagsHas(t *testing.T) {
	// Zero (SectionAll) should have all flags
	var all SectionFlags = 0
	assert.True(t, all.Has(SectionOpcodes))
	assert.True(t, all.Has(SectionStore))
	assert.True(t, all.Has(SectionNative))
	assert.True(t, all.Has(SectionHotSpots))

	// Single flag
	assert.True(t, SectionOpcodes.Has(SectionOpcodes))
	assert.False(t, SectionOpcodes.Has(SectionStore))

	// Combined flags
	combined := SectionOpcodes | SectionStore
	assert.True(t, combined.Has(SectionOpcodes))
	assert.True(t, combined.Has(SectionStore))
	assert.False(t, combined.Has(SectionNative))
}

func TestParseSectionFlags(t *testing.T) {
	tests := []struct {
		input   string
		want    SectionFlags
		wantErr bool
	}{
		{"", 0, false},
		{"all", 0, false},
		{"opcodes", SectionOpcodes, false},
		{"store", SectionStore, false},
		{"native", SectionNative, false},
		{"hotspots", SectionHotSpots, false},
		{"opcodes,store", SectionOpcodes | SectionStore, false},
		{"opcodes,native,hotspots", SectionOpcodes | SectionNative | SectionHotSpots, false},
		{"OPCODES", SectionOpcodes, false},                       // case insensitive
		{"opcodes, store", SectionOpcodes | SectionStore, false}, // spaces trimmed
		{"invalid", 0, true},
		{"opcodes,invalid", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseSectionFlags(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}
