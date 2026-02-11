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

// spinBriefly provides a small measurable delay without using time.Sleep.
// This is more deterministic for timing tests than sleep, which can be flaky.
func spinBriefly() {
	// Spin for a small amount of time to create measurable duration
	start := time.Now()
	for time.Since(start) < 10*time.Microsecond {
		// busy wait
	}
}

func TestProfilerLifecycle(t *testing.T) {
	p := New()

	// Should start in Idle state
	require.Equal(t, StateIdle, p.State())

	// Start should transition to Running
	p.Start()
	require.Equal(t, StateRunning, p.State())

	// Do some work (no sleep needed - just verify state transitions)
	p.BeginOp(OpAdd, OpContext{})
	p.EndOp()

	// Stop should transition back to Idle and return results
	results := p.Stop()
	require.Equal(t, StateIdle, p.State())
	require.NotNil(t, results)
	// Duration may be very small but should exist
	require.NotNil(t, results.OpStats["OpAdd"])

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
		"benchops: Start: profiler is already running (concurrent access or missing Stop)",
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
					"benchops: Start: profiler is already running (concurrent access or missing Stop)",
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
		"benchops: Stop: profiler is not running (missing Start)",
		func() { p.Stop() })
}

func TestProfilerResetPanicsWhenRunning(t *testing.T) {
	p := New()
	p.Start()
	defer p.Stop() // Clean up

	require.PanicsWithValue(t,
		"benchops: Reset: profiler is running (use Stop instead)",
		func() { p.Reset() })
}

func TestOpMeasurement(t *testing.T) {
	p := New()
	// timingEnabled is true by default in New()
	p.Start()

	// Measure some ops (spin loop provides measurable duration without flaky sleep)
	for i := 0; i < 10; i++ {
		p.BeginOp(OpAdd, OpContext{})
		spinBriefly()
		p.EndOp()
	}

	results := p.Stop()

	stat, ok := results.OpStats["OpAdd"]
	require.True(t, ok, "expected OpAdd in results")
	assert.Equal(t, int64(10), stat.Count)
	// Verify timing was recorded (value depends on system speed, just check non-zero)
	assert.NotZero(t, stat.TotalNs, "expected non-zero total duration when timing enabled")
}

// TestNestedStoreCalls verifies that nested store operations correctly pause and
// resume opcode timing.
func TestNestedStoreCalls(t *testing.T) {
	p := New()
	p.Start()

	// Start an opcode
	p.BeginOp(OpCall, OpContext{})

	// Nested store calls should pause opcode timing
	p.BeginStore(StoreGetPackage)

	// Second level nesting
	p.BeginStore(StoreGetObject)
	p.EndStore(100)

	// Third level nesting
	p.BeginStore(StoreGetPackageRealm)
	p.EndStore(50)

	p.EndStore(200)

	// Resume and finish opcode
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
	p.BeginOp(OpCall, OpContext{})
	p.BeginStore(StoreGetPackage)
	p.BeginStore(StoreGetObject)

	// Simulate panic recovery
	p.Recovery()

	// Should be able to continue measuring
	p.BeginOp(OpAdd, OpContext{})
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

	p.BeginOp(OpAdd, OpContext{
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

	p.BeginOp(OpAdd, OpContext{})
	p.EndOp()
	p.BeginStore(StoreGetObject)
	p.EndStore(42)

	results := p.Stop()

	var buf bytes.Buffer
	err := results.WriteReportN(&buf, 10)
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
		"benchops: EndOp: no matching BeginOp",
		func() { p.EndOp() })
}

func TestEndStorePanicsWithoutBegin(t *testing.T) {
	p := New()
	p.Start()
	defer p.Stop()

	require.PanicsWithValue(t,
		"benchops: EndStore: no matching BeginStore",
		func() { p.EndStore(0) })
}

func TestEndNativePanicsWithoutBegin(t *testing.T) {
	p := New()
	p.Start()
	defer p.Stop()

	require.PanicsWithValue(t,
		"benchops: endNative: no matching TraceNative",
		func() { p.endNative() })
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
		p.BeginOp(OpAdd, OpContext{
			File:     "test.gno",
			Line:     10,
			FuncName: "add",
			PkgPath:  "gno.land/r/demo/test",
		})
		p.EndOp()
	}

	for i := 0; i < 3; i++ {
		p.BeginOp(OpMul, OpContext{
			File:     "test.gno",
			Line:     20,
			FuncName: "mul",
			PkgPath:  "gno.land/r/demo/test",
		})
		p.EndOp()
	}

	results := p.Stop()

	// Check location stats are present
	require.NotNil(t, results.LocationStats)
	require.Len(t, results.LocationStats, 2)

	// Check they're sorted by gas (descending)
	assert.GreaterOrEqual(t, results.LocationStats[0].Gas, results.LocationStats[1].Gas)

	// Check line 10 stats (5 OpAdd = 5 * 18 = 90 gas)
	var line10Stat, line20Stat *LocationStat
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

	// Measure ops without setting context (empty OpContext)
	p.BeginOp(OpAdd, OpContext{})
	p.EndOp()

	results := p.Stop()

	// No location stats should be recorded
	assert.Nil(t, results.LocationStats)

	// But op stats should be present
	assert.NotNil(t, results.OpStats["OpAdd"])
}

func TestLocationTrackingReport(t *testing.T) {
	p := New()
	p.Start()

	p.BeginOp(OpAdd, OpContext{
		File:     "counter.gno",
		Line:     15,
		FuncName: "Inc",
		PkgPath:  "gno.land/r/demo/counter",
	})
	p.EndOp()

	results := p.Stop()

	var buf bytes.Buffer
	err := results.WriteReportN(&buf, 10)
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
	p.BeginOp(OpAdd, OpContext{
		File:     "test.gno",
		Line:     10,
		FuncName: "add",
		PkgPath:  "gno.land/r/demo/test",
	})
	p.EndOp()

	p.BeginOp(OpMul, OpContext{
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
	p.TraceNative(NativePrint)()

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
	assert.Contains(t, output, "StoreGetObject: count=1 bytes_read=42 bytes_written=0")

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

	p.BeginOp(OpAdd, OpContext{})
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

func TestIsJSONFormat(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"profile.json", true},
		{"profile.jsonl", true},
		{"profile.JSON", true},  // case insensitive
		{"profile.JSONL", true}, // case insensitive
		{"profile.pprof", false},
		{"profile.pb.gz", false},
		{"profile", false},
		{"/path/to/profile.json", true},
		{"/path/to/profile.pprof", false},
	}

	for _, tc := range tests {
		t.Run(tc.filename, func(t *testing.T) {
			got := IsJSONFormat(tc.filename)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---- Sub-operation (Level 2) tests

func TestSubOpMeasurement(t *testing.T) {
	p := New()
	p.Start()

	// Test basic sub-op tracking
	p.BeginSubOp(SubOpDefineVar, SubOpContext{Line: 10, VarName: "x", File: "test.gno"})
	p.EndSubOp()

	results := p.Stop()
	require.NotNil(t, results.SubOpStats)
	require.Contains(t, results.SubOpStats, "DefineVar")
	assert.Equal(t, int64(1), results.SubOpStats["DefineVar"].Count)
}

func TestSubOpAlwaysEnabled(t *testing.T) {
	p := New()
	p.Start()

	p.BeginSubOp(SubOpDefineVar, SubOpContext{})
	p.EndSubOp()

	results := p.Stop()
	require.NotNil(t, results.SubOpStats)
	assert.Equal(t, int64(1), results.SubOpStats["DefineVar"].Count)
}

func TestSubOpMultipleTypes(t *testing.T) {
	p := New()
	p.Start()

	// Test multiple sub-op types
	for i := 0; i < 3; i++ {
		p.BeginSubOp(SubOpDefineVar, SubOpContext{})
		p.EndSubOp()
	}
	for i := 0; i < 2; i++ {
		p.BeginSubOp(SubOpAssignVar, SubOpContext{})
		p.EndSubOp()
	}

	results := p.Stop()
	assert.Equal(t, int64(3), results.SubOpStats["DefineVar"].Count)
	assert.Equal(t, int64(2), results.SubOpStats["AssignVar"].Count)
}

func TestSubOpString(t *testing.T) {
	tests := map[string]struct {
		op   SubOp
		want string
	}{
		"define":  {op: SubOpDefineVar, want: "DefineVar"},
		"assign":  {op: SubOpAssignVar, want: "AssignVar"},
		"range":   {op: SubOpRangeKey, want: "RangeKey"},
		"unknown": {op: SubOp(0xFE), want: "Unknown"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.op.String())
		})
	}
}

func TestEndSubOpWithoutBeginNoOp(t *testing.T) {
	p := New()
	p.Start()
	defer p.Stop()

	// EndSubOp without BeginSubOp should not panic - it's a no-op
	require.NotPanics(t, func() { p.EndSubOp() })
}

func TestSubOpWithContext(t *testing.T) {
	p := New()
	p.Start()

	// Track variable assignments with context
	p.BeginSubOp(SubOpDefineVar, SubOpContext{
		File:    "test.gno",
		Line:    10,
		VarName: "x",
	})
	p.EndSubOp()

	p.BeginSubOp(SubOpAssignVar, SubOpContext{
		File:    "test.gno",
		Line:    15,
		VarName: "y",
	})
	p.EndSubOp()

	results := p.Stop()

	// Verify sub-op stats
	require.NotNil(t, results.SubOpStats)
	assert.Equal(t, int64(1), results.SubOpStats["DefineVar"].Count)
	assert.Equal(t, int64(1), results.SubOpStats["AssignVar"].Count)

	// Verify variable stats
	require.NotNil(t, results.VarStats)
	require.Len(t, results.VarStats, 2)
}

func TestSubOpIndexedContext(t *testing.T) {
	p := New()
	p.Start()

	// Test indexed operations (e.g., range loop indices)
	p.BeginSubOp(SubOpRangeKey, SubOpContext{
		File:  "test.gno",
		Line:  20,
		Index: 0,
	})
	p.EndSubOp()

	p.BeginSubOp(SubOpRangeValue, SubOpContext{
		File:  "test.gno",
		Line:  20,
		Index: 1,
	})
	p.EndSubOp()

	results := p.Stop()

	require.NotNil(t, results.SubOpStats)
	assert.Equal(t, int64(1), results.SubOpStats["RangeKey"].Count)
	assert.Equal(t, int64(1), results.SubOpStats["RangeValue"].Count)
}

func TestWriteGoldenWithSubOps(t *testing.T) {
	p := New()
	p.Start()

	p.BeginSubOp(SubOpDefineVar, SubOpContext{Line: 10, VarName: "x", File: "test.gno"})
	p.EndSubOp()

	p.BeginSubOp(SubOpAssignVar, SubOpContext{})
	p.EndSubOp()

	results := p.Stop()

	var buf bytes.Buffer
	results.WriteGolden(&buf, SectionSubOps)
	output := buf.String()

	assert.Contains(t, output, "SubOps:")
	assert.Contains(t, output, "AssignVar: count=1")
	assert.Contains(t, output, "DefineVar: count=1")
}

func TestWriteGoldenWithVars(t *testing.T) {
	p := New()
	p.Start()

	p.BeginSubOp(SubOpDefineVar, SubOpContext{
		File:    "test.gno",
		Line:    10,
		VarName: "myVar",
	})
	p.EndSubOp()

	results := p.Stop()

	var buf bytes.Buffer
	results.WriteGolden(&buf, SectionVars)
	output := buf.String()

	assert.Contains(t, output, "Variables:")
	assert.Contains(t, output, "test.gno:10 myVar: count=1")
}

func TestRecoveryResetsSubOps(t *testing.T) {
	p := New()
	p.Start()

	// Start a sub-op without ending it
	p.BeginSubOp(SubOpDefineVar, SubOpContext{})

	// Simulate panic recovery
	p.Recovery()

	// Should be able to continue measuring
	p.BeginSubOp(SubOpAssignVar, SubOpContext{})
	p.EndSubOp()

	results := p.Stop()

	// Only AssignVar should be in results (DefineVar was not ended)
	assert.NotContains(t, results.SubOpStats, "DefineVar")
	assert.Contains(t, results.SubOpStats, "AssignVar")
}

// ---- Call stack tracking tests

func TestStackTrackingBasic(t *testing.T) {
	p := New()
	// stackEnabled is true by default in New()
	p.Start()

	// Simulate a call stack: main -> handleRequest -> computeValue
	p.PushCall("main", "gno.land/r/demo/test", "test.gno", 5)
	p.PushCall("handleRequest", "gno.land/r/demo/test", "test.gno", 10)
	p.PushCall("computeValue", "gno.land/r/demo/test", "test.gno", 20)

	// Record an op while in computeValue
	p.BeginOp(OpAdd, OpContext{File: "test.gno", Line: 25})
	p.EndOp()

	// Pop back to handleRequest
	p.PopCall()

	// Record another op while in handleRequest
	p.BeginOp(OpMul, OpContext{File: "test.gno", Line: 12})
	p.EndOp()

	p.PopCall() // back to main
	p.PopCall() // exit main

	results := p.Stop()

	// Check that we have stack samples
	require.NotNil(t, results.StackSamples)
	require.GreaterOrEqual(t, len(results.StackSamples), 2)

	// Check the first sample has the full call stack (leaf-to-root)
	found3Deep := false
	found2Deep := false
	for _, sample := range results.StackSamples {
		if len(sample.Stack) == 3 {
			found3Deep = true
			assert.Equal(t, "computeValue", sample.Stack[0].Func)
			assert.Equal(t, "handleRequest", sample.Stack[1].Func)
			assert.Equal(t, "main", sample.Stack[2].Func)
		}
		if len(sample.Stack) == 2 {
			found2Deep = true
			assert.Equal(t, "handleRequest", sample.Stack[0].Func)
			assert.Equal(t, "main", sample.Stack[1].Func)
		}
	}
	assert.True(t, found3Deep, "expected a sample with 3-level call stack")
	assert.True(t, found2Deep, "expected a sample with 2-level call stack")
}

func TestStackTrackingCanBeDisabled(t *testing.T) {
	p := New()
	// Internal test: directly set field to test disabled behavior.
	// Production code uses WithoutStacks() option via Start().
	p.stackEnabled = false
	p.Start()

	p.PushCall("main", "gno.land/r/demo/test", "test.gno", 5)
	p.BeginOp(OpAdd, OpContext{File: "test.gno", Line: 10})
	p.EndOp()
	p.PopCall()

	results := p.Stop()

	// No stack samples should be recorded when disabled
	assert.Empty(t, results.StackSamples)
}

func TestRecoveryResetsCallStack(t *testing.T) {
	p := New()
	// stackEnabled is true by default in New()
	p.Start()

	// Push some calls without popping
	p.PushCall("main", "pkg", "test.gno", 1)
	p.PushCall("nested", "pkg", "test.gno", 10)

	// Simulate panic recovery
	p.Recovery()

	// Call stack should be empty now
	// Test by pushing a new call and recording an op
	p.PushCall("newMain", "pkg", "test.gno", 1)
	p.BeginOp(OpAdd, OpContext{File: "test.gno", Line: 5})
	p.EndOp()
	p.PopCall()

	results := p.Stop()

	// Only the newMain call should be in the stack samples
	require.NotNil(t, results.StackSamples)
	for _, sample := range results.StackSamples {
		// Should only have 1 frame (newMain)
		assert.Equal(t, 1, len(sample.Stack))
		assert.Equal(t, "newMain", sample.Stack[0].Func)
	}
}

func TestStackTrackingPprofOutput(t *testing.T) {
	if !Enabled {
		t.Skip("requires gnobench build tag")
	}
	p := New()
	// stackEnabled is true by default in New()
	p.Start()

	// Simulate a call stack
	p.PushCall("main", "gno.land/r/demo/test", "test.gno", 5)
	p.PushCall("compute", "gno.land/r/demo/test", "test.gno", 10)

	// Record some ops
	for i := 0; i < 3; i++ {
		p.BeginOp(OpAdd, OpContext{File: "test.gno", Line: 15})
		p.EndOp()
	}

	p.PopCall()
	p.PopCall()

	results := p.Stop()

	// Write pprof output
	var buf bytes.Buffer
	err := results.WritePprof(&buf)
	require.NoError(t, err)
	assert.NotZero(t, buf.Len())
}

// ---- Global Start() with options integration tests

func TestGlobalStartWithoutTiming(t *testing.T) {
	if !Enabled {
		t.Skip("requires gnobench build tag")
	}
	Start(WithoutTiming())
	defer func() {
		if IsRunning() {
			Stop()
		}
	}()

	R().BeginOp(OpAdd, OpContext{})
	R().EndOp()
	results := Stop()

	// Verify timing is disabled
	stat := results.OpStats["OpAdd"]
	require.NotNil(t, stat)
	assert.Zero(t, stat.TotalNs, "expected zero timing when WithoutTiming used")
	assert.Equal(t, int64(1), stat.Count)
	assert.False(t, results.TimingEnabled)
}

func TestGlobalStartWithoutStacks(t *testing.T) {
	if !Enabled {
		t.Skip("requires gnobench build tag")
	}
	Start(WithoutStacks())
	defer func() {
		if IsRunning() {
			Stop()
		}
	}()

	R().PushCall("main", "gno.land/r/demo/test", "test.gno", 1)
	R().BeginOp(OpAdd, OpContext{File: "test.gno", Line: 5})
	R().EndOp()
	R().PopCall()
	results := Stop()

	// Verify stacks are disabled
	assert.Empty(t, results.StackSamples, "expected no stack samples when WithoutStacks used")
	// But ops should still be recorded
	assert.Equal(t, int64(1), results.OpStats["OpAdd"].Count)
}

func TestGlobalStartWithBothOptionsDisabled(t *testing.T) {
	if !Enabled {
		t.Skip("requires gnobench build tag")
	}
	Start(WithoutTiming(), WithoutStacks())
	defer func() {
		if IsRunning() {
			Stop()
		}
	}()

	R().PushCall("main", "gno.land/r/demo/test", "test.gno", 1)
	R().BeginOp(OpAdd, OpContext{File: "test.gno", Line: 5})
	R().EndOp()
	R().PopCall()
	results := Stop()

	// Verify both are disabled
	assert.Zero(t, results.OpStats["OpAdd"].TotalNs, "expected zero timing")
	assert.Empty(t, results.StackSamples, "expected no stack samples")
	assert.False(t, results.TimingEnabled)

	// But counts should still work
	assert.Equal(t, int64(1), results.OpStats["OpAdd"].Count)
}

// ---- Store I/O byte tracking tests

func TestStoreStatBytesRead(t *testing.T) {
	var stat StoreStat
	stat.Record(StoreGetObject, 1024, time.Millisecond)
	stat.Record(StoreGetObject, 2048, time.Millisecond)

	assert.Equal(t, int64(3072), stat.BytesRead)
	assert.Equal(t, int64(0), stat.BytesWritten)
	assert.Equal(t, int64(3072), stat.TotalSize) // backward compat
}

func TestStoreStatBytesWritten(t *testing.T) {
	var stat StoreStat
	stat.Record(StoreSetObject, 512, time.Millisecond)

	assert.Equal(t, int64(0), stat.BytesRead)
	assert.Equal(t, int64(512), stat.BytesWritten)
	assert.Equal(t, int64(512), stat.TotalSize) // backward compat
}

func TestStoreStatMixedOperations(t *testing.T) {
	var stat StoreStat
	// Read operations
	stat.Record(StoreGetObject, 100, time.Millisecond)
	stat.Record(StoreGetPackage, 200, time.Millisecond)
	// Write operations
	stat.Record(StoreSetObject, 50, time.Millisecond)
	stat.Record(StoreSetPackage, 75, time.Millisecond)

	assert.Equal(t, int64(300), stat.BytesRead)
	assert.Equal(t, int64(125), stat.BytesWritten)
	assert.Equal(t, int64(425), stat.TotalSize)
}

func TestStoreOpIsRead(t *testing.T) {
	readOps := []StoreOp{
		StoreGetObject, StoreGetPackage, StoreGetType,
		StoreGetBlockNode, StoreGetMemPackage, StoreGetPackageRealm,
		StoreGet, AminoUnmarshal,
	}
	for _, op := range readOps {
		assert.True(t, op.IsRead(), "expected %s to be a read op", op)
		assert.False(t, op.IsWrite(), "expected %s not to be a write op", op)
	}
}

func TestStoreOpIsWrite(t *testing.T) {
	writeOps := []StoreOp{
		StoreSetObject, StoreSetPackage, StoreSetType,
		StoreSetBlockNode, StoreAddMemPackage, StoreSetPackageRealm,
		StoreSet, AminoMarshal, AminoMarshalAny, FinalizeTx,
	}
	for _, op := range writeOps {
		assert.True(t, op.IsWrite(), "expected %s to be a write op", op)
		assert.False(t, op.IsRead(), "expected %s not to be a read op", op)
	}
}

func TestStoreOpDeleteIsNeither(t *testing.T) {
	// Delete operations don't count as read or write
	assert.False(t, StoreDeleteObject.IsRead())
	assert.False(t, StoreDeleteObject.IsWrite())
}

func TestProfilerBytesTracking(t *testing.T) {
	p := New()
	p.Start()

	// Track read operation
	p.BeginStore(StoreGetObject)
	p.EndStore(1024)

	// Track write operation
	p.BeginStore(StoreSetObject)
	p.EndStore(512)

	results := p.Stop()

	// Check read bytes
	getStat := results.StoreStats["StoreGetObject"]
	require.NotNil(t, getStat)
	assert.Equal(t, int64(1), getStat.Count)
	assert.Equal(t, int64(1024), getStat.BytesRead)
	assert.Equal(t, int64(0), getStat.BytesWritten)

	// Check write bytes
	setStat := results.StoreStats["StoreSetObject"]
	require.NotNil(t, setStat)
	assert.Equal(t, int64(1), setStat.Count)
	assert.Equal(t, int64(0), setStat.BytesRead)
	assert.Equal(t, int64(512), setStat.BytesWritten)
}

func TestWriteGoldenWithStoreBytes(t *testing.T) {
	p := New()
	p.Start()

	p.BeginStore(StoreGetObject)
	p.EndStore(1024)

	p.BeginStore(StoreSetObject)
	p.EndStore(512)

	results := p.Stop()

	var buf bytes.Buffer
	results.WriteGolden(&buf, SectionStore)
	output := buf.String()

	assert.Contains(t, output, "Store:")
	assert.Contains(t, output, "StoreGetObject: count=1 bytes_read=1024 bytes_written=0")
	assert.Contains(t, output, "StoreSetObject: count=1 bytes_read=0 bytes_written=512")
}

func TestMergeResultsBytesTracking(t *testing.T) {
	// Create first result set
	p1 := New()
	p1.Start()
	p1.BeginStore(StoreGetObject)
	p1.EndStore(1000)
	results1 := p1.Stop()

	// Create second result set
	p2 := New()
	p2.Start()
	p2.BeginStore(StoreGetObject)
	p2.EndStore(500)
	p2.BeginStore(StoreSetObject)
	p2.EndStore(200)
	results2 := p2.Stop()

	// Merge results
	merged := MergeResults(results1, results2)

	// Check merged stats
	getStat := merged.StoreStats["StoreGetObject"]
	require.NotNil(t, getStat)
	assert.Equal(t, int64(2), getStat.Count)
	assert.Equal(t, int64(1500), getStat.BytesRead) // 1000 + 500

	setStat := merged.StoreStats["StoreSetObject"]
	require.NotNil(t, setStat)
	assert.Equal(t, int64(1), setStat.Count)
	assert.Equal(t, int64(200), setStat.BytesWritten)
}
