package benchops

import (
	"testing"
)

// --- OpCode timing ---

func TestSwitchOpCode_Basic(t *testing.T) {
	InitMeasure()
	old := SwitchOpCode(0x01)
	if old != invalidCode {
		t.Fatalf("expected old code %#x, got %#x", invalidCode, old)
	}
	if measure.opCounts[0x01] != 1 {
		t.Fatalf("expected opCounts[0x01] == 1, got %d", measure.opCounts[0x01])
	}
	if measure.curOpCode != 0x01 {
		t.Fatalf("expected curOpCode 0x01, got %#x", measure.curOpCode)
	}
}

func TestSwitchOpCode_Chain(t *testing.T) {
	InitMeasure()

	oldA := SwitchOpCode(0x10) // invalidCode -> A
	if oldA != invalidCode {
		t.Fatalf("first switch should return invalidCode, got %#x", oldA)
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

	for _, code := range []byte{0x10, 0x20, 0x30} {
		if measure.opCounts[code] != 1 {
			t.Errorf("opCounts[%#x] = %d, want 1", code, measure.opCounts[code])
		}
		if measure.opAccumDur[code] <= 0 {
			t.Errorf("opAccumDur[%#x] = %v, want > 0", code, measure.opAccumDur[code])
		}
	}
}

func TestSwitchOpCode_InvalidCode_Panics(t *testing.T) {
	InitMeasure()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalidCode, got none")
		}
	}()
	SwitchOpCode(invalidCode)
}

func TestResumeOpCode(t *testing.T) {
	InitMeasure()

	SwitchOpCode(0x01)              // start op A (count=1)
	old := StartStore(0x10)         // pause op A, start store
	if old != 0x01 {
		t.Fatalf("expected old 0x01, got %#x", old)
	}
	if measure.curOpCode != invalidCode {
		t.Fatalf("curOpCode should be invalidCode during store, got %#x", measure.curOpCode)
	}

	ResumeOpCode(old) // resume op A — count should stay 1

	if measure.opCounts[0x01] != 1 {
		t.Fatalf("resume should not increment count; got %d", measure.opCounts[0x01])
	}
	if measure.curOpCode != 0x01 {
		t.Fatalf("curOpCode should be 0x01 after resume, got %#x", measure.curOpCode)
	}
}

func TestStopOpCode(t *testing.T) {
	InitMeasure()
	SwitchOpCode(0x05)
	StopOpCode()

	if measure.curOpCode != invalidCode {
		t.Fatalf("curOpCode should be invalidCode after StopOpCode, got %#x", measure.curOpCode)
	}
	if measure.opAccumDur[0x05] <= 0 {
		t.Errorf("opAccumDur[0x05] = %v, want > 0", measure.opAccumDur[0x05])
	}
}

// --- Store timing ---

func TestStartStopStore(t *testing.T) {
	InitMeasure()
	vmOp := byte(0x10)
	storeCode := byte(0x01)

	SwitchOpCode(vmOp)
	old := StartStore(storeCode)
	if old != vmOp {
		t.Fatalf("StartStore should return old VM op %#x, got %#x", vmOp, old)
	}

	StopStore(storeCode, old, 42)

	if measure.storeCounts[storeCode] != 1 {
		t.Fatalf("storeCounts = %d, want 1", measure.storeCounts[storeCode])
	}
	if measure.storeAccumDur[storeCode] <= 0 {
		t.Errorf("storeAccumDur = %v, want > 0", measure.storeAccumDur[storeCode])
	}
	if measure.storeAccumSize[storeCode] != 42 {
		t.Fatalf("storeAccumSize = %d, want 42", measure.storeAccumSize[storeCode])
	}
	if measure.curOpCode != vmOp {
		t.Fatalf("VM op should be resumed; curOpCode = %#x, want %#x", measure.curOpCode, vmOp)
	}
}

func TestStartStore_SuspendsVMOp(t *testing.T) {
	InitMeasure()
	SwitchOpCode(0x10)
	old := StartStore(0x01)

	// During store operation, VM timeline should be paused.
	if measure.curOpCode != invalidCode {
		t.Fatalf("curOpCode should be invalidCode during store, got %#x", measure.curOpCode)
	}

	StopStore(0x01, old, 0) // clean up
}

// --- Native timing ---

func TestStartStopNative(t *testing.T) {
	InitMeasure()
	vmOp := byte(0x10)
	nativeCode := byte(0x01)

	SwitchOpCode(vmOp)
	old := StartNative(nativeCode)
	if old != vmOp {
		t.Fatalf("StartNative should return old VM op %#x, got %#x", vmOp, old)
	}

	StopNative(nativeCode, old)

	if measure.nativeCounts[nativeCode] != 1 {
		t.Fatalf("nativeCounts = %d, want 1", measure.nativeCounts[nativeCode])
	}
	if measure.nativeAccumDur[nativeCode] <= 0 {
		t.Errorf("nativeAccumDur = %v, want > 0", measure.nativeAccumDur[nativeCode])
	}
	if measure.curOpCode != vmOp {
		t.Fatalf("VM op should be resumed; curOpCode = %#x, want %#x", measure.curOpCode, vmOp)
	}
}

func TestStartNative_InvalidCode_Panics(t *testing.T) {
	InitMeasure()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalidCode, got none")
		}
	}()
	StartNative(invalidCode)
}

// --- InitMeasure ---

func TestInitMeasure_Resets(t *testing.T) {
	// Accumulate some state.
	InitMeasure()
	SwitchOpCode(0x01)
	SwitchOpCode(0x02)
	old := StartStore(0x03)
	StopStore(0x03, old, 100)
	nOld := StartNative(0x04)
	StopNative(0x04, nOld)
	StopOpCode()

	// Reset.
	InitMeasure()

	if measure.curOpCode != invalidCode {
		t.Fatalf("curOpCode should be invalidCode after reset, got %#x", measure.curOpCode)
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
