package benchops

import (
	"testing"
	"time"
)

// --- OpCode timing ---

func TestSwitchOpCode_Basic(t *testing.T) {
	InitMeasure()
	old := SwitchOpCode(0x01)
	if old != CPUOpInvalid {
		t.Fatalf("expected old code %#x, got %#x", CPUOpInvalid, old)
	}
	if measure.opCounts[0x01] != 1 {
		t.Fatalf("expected opCounts[0x01] == 1, got %d", measure.opCounts[0x01])
	}
	if measure.curCPUOp != 0x01 {
		t.Fatalf("expected curCPUOp 0x01, got %#x", measure.curCPUOp)
	}
}

func TestSwitchOpCode_Chain(t *testing.T) {
	InitMeasure()

	oldA := SwitchOpCode(0x10) // invalid -> A
	if oldA != CPUOpInvalid {
		t.Fatalf("first switch should return CPUOpInvalid, got %#x", oldA)
	}

	oldB := SwitchOpCode(0x20) // A -> B
	if oldB != 0x10 {
		t.Fatalf("expected old 0x10, got %#x", oldB)
	}

	oldC := SwitchOpCode(0x30) // B -> C
	if oldC != 0x20 {
		t.Fatalf("expected old 0x20, got %#x", oldC)
	}

	StopOpCode() // finalize C

	for _, code := range []CPUOp{0x10, 0x20, 0x30} {
		if measure.opCounts[byte(code)] != 1 {
			t.Errorf("opCounts[%#x] = %d, want 1", code, measure.opCounts[byte(code)])
		}
		if measure.opAccumDur[byte(code)] <= 0 {
			t.Errorf("opAccumDur[%#x] = %v, want > 0", code, measure.opAccumDur[byte(code)])
		}
	}
}

func TestSwitchOpCode_InvalidCode_Panics(t *testing.T) {
	InitMeasure()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for CPUOpInvalid, got none")
		}
	}()
	SwitchOpCode(CPUOpInvalid)
}

func TestResumeOpCode(t *testing.T) {
	InitMeasure()

	SwitchOpCode(0x01)        // start A (count=1)
	old := SwitchOpCode(0x02) // A -> B
	if old != 0x01 {
		t.Fatalf("expected old 0x01, got %#x", old)
	}

	ResumeOpCode(0x01) // resume A — count should stay 1

	if measure.opCounts[0x01] != 1 {
		t.Fatalf("resume should not increment count; got %d", measure.opCounts[0x01])
	}
	if measure.curCPUOp != 0x01 {
		t.Fatalf("curCPUOp should be 0x01 after resume, got %#x", measure.curCPUOp)
	}
}

func TestStopOpCode(t *testing.T) {
	InitMeasure()
	SwitchOpCode(0x05)
	StopOpCode()

	if measure.curCPUOp != CPUOpInvalid {
		t.Fatalf("curCPUOp should be CPUOpInvalid after StopOpCode, got %#x", measure.curCPUOp)
	}
	if measure.opAccumDur[0x05] <= 0 {
		t.Errorf("opAccumDur[0x05] = %v, want > 0", measure.opAccumDur[0x05])
	}
}

// --- Store timing ---

func TestStartStopStore(t *testing.T) {
	InitMeasure()
	vmOp := CPUOp(0x10)
	storeCode := StoreOp(0x01)

	SwitchOpCode(vmOp)
	oldCPU, oldStore := StartStore(storeCode)
	if oldCPU != vmOp {
		t.Fatalf("StartStore should return old VM op %#x, got %#x", vmOp, oldCPU)
	}
	if oldStore != StoreOpInvalid {
		t.Fatalf("StartStore should return StoreOpInvalid, got %#x", oldStore)
	}

	StopStore(storeCode, oldCPU, oldStore, 42)

	if measure.storeCounts[byte(storeCode)] != 1 {
		t.Fatalf("storeCounts = %d, want 1", measure.storeCounts[byte(storeCode)])
	}
	if measure.storeAccumDur[byte(storeCode)] <= 0 {
		t.Errorf("storeAccumDur = %v, want > 0", measure.storeAccumDur[byte(storeCode)])
	}
	if measure.storeAccumSize[byte(storeCode)] != 42 {
		t.Fatalf("storeAccumSize = %d, want 42", measure.storeAccumSize[byte(storeCode)])
	}
	if measure.curCPUOp != vmOp {
		t.Fatalf("VM op should be resumed; curCPUOp = %#x, want %#x", measure.curCPUOp, vmOp)
	}
}

func TestStartStore_SuspendsVMOp(t *testing.T) {
	InitMeasure()
	SwitchOpCode(0x10)
	oldCPU, oldStore := StartStore(0x01)

	// During store operation, VM timeline should be paused.
	if measure.curCPUOp != CPUOpInvalid {
		t.Fatalf("curCPUOp should be CPUOpInvalid during store, got %#x", measure.curCPUOp)
	}

	StopStore(0x01, oldCPU, oldStore, 0) // clean up
}

// --- Nested store timing (regression test for nesting bug) ---

func TestNestedStoreOps(t *testing.T) {
	InitMeasure()
	vmOp := CPUOp(0x10)
	outerStore := StoreOp(0x01)
	innerStore := StoreOp(0x02)

	SwitchOpCode(vmOp)

	// Outer store op (e.g. RealmFinalizeTx)
	outerCPU, outerStoreOld := StartStore(outerStore)
	time.Sleep(1 * time.Millisecond) // outer's own work

	// Inner store op (e.g. SetObject called from within FinalizeRealmTransaction)
	innerCPU, innerStoreOld := StartStore(innerStore)
	time.Sleep(1 * time.Millisecond) // inner's work
	StopStore(innerStore, innerCPU, innerStoreOld, 10)

	time.Sleep(1 * time.Millisecond) // more outer work
	StopStore(outerStore, outerCPU, outerStoreOld, 20)

	// Both store ops should have accumulated time.
	if measure.storeAccumDur[byte(outerStore)] <= 0 {
		t.Errorf("outer store op duration = %v, want > 0", measure.storeAccumDur[byte(outerStore)])
	}
	if measure.storeAccumDur[byte(innerStore)] <= 0 {
		t.Errorf("inner store op duration = %v, want > 0", measure.storeAccumDur[byte(innerStore)])
	}
	// Inner should have saved/restored outer store op, not CPU op.
	if innerCPU != CPUOpInvalid {
		t.Fatalf("inner StartStore should save CPUOpInvalid, got %#x", innerCPU)
	}
	if innerStoreOld != outerStore {
		t.Fatalf("inner StartStore should save outer store op %#x, got %#x", outerStore, innerStoreOld)
	}
	// VM op should be restored after everything.
	if measure.curCPUOp != vmOp {
		t.Fatalf("VM op should be restored; curCPUOp = %#x, want %#x", measure.curCPUOp, vmOp)
	}
}

// --- Native timing ---

func TestStartStopNative(t *testing.T) {
	InitMeasure()
	vmOp := CPUOp(0x10)
	nativeCode := NativeOp(0x01)

	SwitchOpCode(vmOp)
	old := StartNative(nativeCode)
	if old != vmOp {
		t.Fatalf("StartNative should return old VM op %#x, got %#x", vmOp, old)
	}

	StopNative(nativeCode, old)

	if measure.nativeCounts[byte(nativeCode)] != 1 {
		t.Fatalf("nativeCounts = %d, want 1", measure.nativeCounts[byte(nativeCode)])
	}
	if measure.nativeAccumDur[byte(nativeCode)] <= 0 {
		t.Errorf("nativeAccumDur = %v, want > 0", measure.nativeAccumDur[byte(nativeCode)])
	}
	if measure.curCPUOp != vmOp {
		t.Fatalf("VM op should be resumed; curCPUOp = %#x, want %#x", measure.curCPUOp, vmOp)
	}
}

func TestStartNative_InvalidCode_Panics(t *testing.T) {
	InitMeasure()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for NativeOpInvalid, got none")
		}
	}()
	StartNative(NativeOpInvalid)
}

// --- InitMeasure ---

func TestInitMeasure_Resets(t *testing.T) {
	// Accumulate some state.
	InitMeasure()
	SwitchOpCode(0x01)
	SwitchOpCode(0x02)
	oldCPU, oldStore := StartStore(0x03)
	StopStore(0x03, oldCPU, oldStore, 100)
	nOld := StartNative(0x04)
	StopNative(0x04, nOld)
	StopOpCode()

	// Reset.
	InitMeasure()

	if measure.curCPUOp != CPUOpInvalid {
		t.Fatalf("curCPUOp should be CPUOpInvalid after reset, got %#x", measure.curCPUOp)
	}
	for i := range 256 {
		if measure.opCounts[i] != 0 {
			t.Fatalf("opCounts[%d] = %d after reset", i, measure.opCounts[i])
		}
		if measure.opAccumDur[i] != 0 {
			t.Fatalf("opAccumDur[%d] = %v after reset", i, measure.opAccumDur[i])
		}
		if measure.storeCounts[i] != 0 {
			t.Fatalf("storeCounts[%d] = %d after reset", i, measure.storeCounts[i])
		}
		if measure.storeAccumDur[i] != 0 {
			t.Fatalf("storeAccumDur[%d] = %v after reset", i, measure.storeAccumDur[i])
		}
		if measure.storeAccumSize[i] != 0 {
			t.Fatalf("storeAccumSize[%d] = %d after reset", i, measure.storeAccumSize[i])
		}
		if measure.nativeCounts[i] != 0 {
			t.Fatalf("nativeCounts[%d] = %d after reset", i, measure.nativeCounts[i])
		}
		if measure.nativeAccumDur[i] != 0 {
			t.Fatalf("nativeAccumDur[%d] = %v after reset", i, measure.nativeAccumDur[i])
		}
	}
}
