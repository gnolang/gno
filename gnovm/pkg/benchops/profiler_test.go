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

	// Do some work
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
		"benchops: Stop called on non-running profiler",
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
	p.Start()

	// Measure some ops
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
	p.EndOp()

	results := p.Stop()

	var buf bytes.Buffer
	err := results.WriteJSON(&buf)
	require.NoError(t, err)
	assert.NotZero(t, buf.Len(), "expected non-empty JSON output")
	assert.Contains(t, buf.String(), "OpAdd")
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
