package benchops

import (
	"bytes"
	"testing"
	"time"
)

func TestProfilerLifecycle(t *testing.T) {
	p := New(DefaultConfig())

	// Should start in Idle state
	if p.State() != StateIdle {
		t.Fatalf("expected StateIdle, got %v", p.State())
	}

	// Start should transition to Running
	p.Start()
	if p.State() != StateRunning {
		t.Fatalf("expected StateRunning, got %v", p.State())
	}

	// Do some work
	p.BeginOp(OpAdd)
	time.Sleep(time.Microsecond)
	p.EndOp()

	// Stop should transition to Stopped and return results
	results := p.Stop()
	if p.State() != StateStopped {
		t.Fatalf("expected StateStopped, got %v", p.State())
	}
	if results == nil {
		t.Fatal("expected non-nil results")
	}
	if results.Duration == 0 {
		t.Error("expected non-zero duration")
	}

	// Reset should return to Idle
	p.Reset()
	if p.State() != StateIdle {
		t.Fatalf("expected StateIdle after Reset, got %v", p.State())
	}
}

func TestProfilerStartPanics(t *testing.T) {
	p := New(DefaultConfig())
	p.Start()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on double Start")
		}
	}()

	p.Start() // should panic
}

func TestProfilerStopPanics(t *testing.T) {
	p := New(DefaultConfig())

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on Stop before Start")
		}
	}()

	p.Stop() // should panic
}

func TestOpMeasurement(t *testing.T) {
	p := New(DefaultConfig())
	p.Start()

	// Measure some ops
	for i := 0; i < 10; i++ {
		p.BeginOp(OpAdd)
		time.Sleep(10 * time.Microsecond)
		p.EndOp()
	}

	results := p.Stop()

	stat, ok := results.OpStats["OpAdd"]
	if !ok {
		t.Fatal("expected OpAdd in results")
	}
	if stat.Count != 10 {
		t.Errorf("expected count 10, got %d", stat.Count)
	}
	if stat.TotalNs == 0 {
		t.Error("expected non-zero total duration")
	}
	if stat.AvgNs == 0 {
		t.Error("expected non-zero average duration")
	}
}

func TestNestedStoreCalls(t *testing.T) {
	p := New(DefaultConfig())
	p.Start()

	// Start an opcode
	p.BeginOp(OpCall)
	time.Sleep(10 * time.Microsecond)

	// Nested store calls should pause opcode timing
	p.BeginStore(StoreGetPackage)
	time.Sleep(5 * time.Microsecond)

	// Second level nesting
	p.BeginStore(StoreGetObject)
	time.Sleep(5 * time.Microsecond)
	p.EndStore(100)

	// Third level nesting
	p.BeginStore(StoreGetPackageRealm)
	time.Sleep(5 * time.Microsecond)
	p.EndStore(50)

	p.EndStore(200)

	// Resume and finish opcode
	time.Sleep(10 * time.Microsecond)
	p.EndOp()

	results := p.Stop()

	// Check opcode was recorded
	opStat, ok := results.OpStats["OpCall"]
	if !ok {
		t.Fatal("expected OpCall in results")
	}
	if opStat.Count != 1 {
		t.Errorf("expected OpCall count 1, got %d", opStat.Count)
	}

	// Check all store ops were recorded
	stores := []string{"StoreGetPackage", "StoreGetObject", "StoreGetPackageRealm"}
	for _, name := range stores {
		stat, ok := results.StoreStats[name]
		if !ok {
			t.Errorf("expected %s in results", name)
			continue
		}
		if stat.Count != 1 {
			t.Errorf("expected %s count 1, got %d", name, stat.Count)
		}
	}

	// Check sizes were recorded
	if results.StoreStats["StoreGetObject"].TotalSize != 100 {
		t.Errorf("expected StoreGetObject size 100, got %d", results.StoreStats["StoreGetObject"].TotalSize)
	}
}

func TestPanicRecovery(t *testing.T) {
	p := New(DefaultConfig())
	p.Start()

	// Start an op and some store calls
	p.BeginOp(OpCall)
	p.BeginStore(StoreGetPackage)
	p.BeginStore(StoreGetObject)

	// Simulate panic recovery
	p.Recovery()

	// Should be able to continue measuring
	p.BeginOp(OpAdd)
	time.Sleep(time.Microsecond)
	p.EndOp()

	results := p.Stop()

	// Only OpAdd should be in results (OpCall was not ended)
	if _, ok := results.OpStats["OpCall"]; ok {
		t.Error("OpCall should not be in results after recovery")
	}
	if _, ok := results.OpStats["OpAdd"]; !ok {
		t.Error("OpAdd should be in results after recovery")
	}
}

func TestDisabledMeasurements(t *testing.T) {
	cfg := Config{
		EnableOps:    false,
		EnableStore:  false,
		EnableNative: false,
	}
	p := New(cfg)
	p.Start()

	// These should all be no-ops
	p.BeginOp(OpAdd)
	p.EndOp()
	p.BeginStore(StoreGetObject)
	p.EndStore(100)
	p.BeginNative(NativePrint)
	p.EndNative()

	results := p.Stop()

	if len(results.OpStats) != 0 {
		t.Error("expected no op stats when disabled")
	}
	if len(results.StoreStats) != 0 {
		t.Error("expected no store stats when disabled")
	}
	if len(results.NativeStats) != 0 {
		t.Error("expected no native stats when disabled")
	}
}

func TestResultsJSON(t *testing.T) {
	p := New(DefaultConfig())
	p.Start()

	p.BeginOp(OpAdd)
	p.EndOp()

	results := p.Stop()

	var buf bytes.Buffer
	if err := results.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected non-empty JSON output")
	}

	// Basic sanity check
	if !bytes.Contains(buf.Bytes(), []byte("OpAdd")) {
		t.Error("expected OpAdd in JSON output")
	}
}

func TestResultsReport(t *testing.T) {
	p := New(DefaultConfig())
	p.Start()

	p.BeginOp(OpAdd)
	p.EndOp()
	p.BeginStore(StoreGetObject)
	p.EndStore(42)

	results := p.Stop()

	var buf bytes.Buffer
	if err := results.WriteReport(&buf, 10); err != nil {
		t.Fatalf("WriteReport failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected non-empty report output")
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("OpAdd")) {
		t.Error("expected OpAdd in report output")
	}
	if !bytes.Contains([]byte(output), []byte("StoreGetObject")) {
		t.Error("expected StoreGetObject in report output")
	}
}

func TestOpString(t *testing.T) {
	tests := []struct {
		op   Op
		want string
	}{
		{OpAdd, "OpAdd"},
		{OpCall, "OpCall"},
		{Op(0xFE), "OpUnknown"}, // unknown op
	}

	for _, tt := range tests {
		got := tt.op.String()
		if got != tt.want {
			t.Errorf("Op(%#x).String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestStoreOpString(t *testing.T) {
	tests := []struct {
		op   StoreOp
		want string
	}{
		{StoreGetObject, "StoreGetObject"},
		{StoreSetPackage, "StoreSetPackage"},
		{StoreOp(0xFE), "StoreOpUnknown"}, // unknown op
	}

	for _, tt := range tests {
		got := tt.op.String()
		if got != tt.want {
			t.Errorf("StoreOp(%#x).String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestNativeOpString(t *testing.T) {
	tests := []struct {
		op   NativeOp
		want string
	}{
		{NativePrint, "NativePrint"},
		{NativePrint1, "NativePrint1"},
		{NativeOp(0xFE), "NativeOpUnknown"}, // unknown op
	}

	for _, tt := range tests {
		got := tt.op.String()
		if got != tt.want {
			t.Errorf("NativeOp(%#x).String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestGetNativePrintCode(t *testing.T) {
	tests := []struct {
		size int
		want NativeOp
	}{
		{1, NativePrint1},
		{1000, NativePrint1000},
		{10000, NativePrint1e4},
	}

	for _, tt := range tests {
		got := GetNativePrintCode(tt.size)
		if got != tt.want {
			t.Errorf("GetNativePrintCode(%d) = %v, want %v", tt.size, got, tt.want)
		}
	}
}

func TestGetNativePrintCodePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid print size")
		}
	}()

	GetNativePrintCode(42) // should panic
}

func TestGetOpGas(t *testing.T) {
	// Test known op
	if gas := GetOpGas(OpAdd); gas != 18 {
		t.Errorf("GetOpGas(OpAdd) = %d, want 18", gas)
	}

	// Test unknown op returns 1
	if gas := GetOpGas(Op(0xFE)); gas != 1 {
		t.Errorf("GetOpGas(unknown) = %d, want 1", gas)
	}
}
