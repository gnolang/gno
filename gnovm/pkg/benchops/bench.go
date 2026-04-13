package benchops

import (
	"runtime"
	"time"
)

var measure bench

type bench struct {
	// Current active op — at most one is non-invalid at a time.
	curCPUOp    CPUOp
	curStoreOp  StoreOp
	curNativeOp NativeOp
	curStart    time.Time
	timeZero    time.Time

	// Opcode timing: single timeline, always one op active.
	opCounts   [256]int64
	opAccumDur [256]time.Duration

	// Store timing: own accumulators.
	storeCounts    [256]int64
	storeAccumDur  [256]time.Duration
	storeAccumSize [256]int64

	// Native timing.
	nativeCounts   [256]int64
	nativeAccumDur [256]time.Duration
}

func InitMeasure() {
	measure = bench{
		curCPUOp:    CPUOpInvalid,
		curStoreOp:  StoreOpInvalid,
		curNativeOp: NativeOpInvalid,
	}
}

// reapTime attributes elapsed time since curStart to whichever
// op is currently active, then sets curStart = now.
func reapTime() {
	now := time.Now()
	if measure.curStart != measure.timeZero {
		elapsed := now.Sub(measure.curStart)
		if measure.curCPUOp != CPUOpInvalid {
			measure.opAccumDur[byte(measure.curCPUOp)] += elapsed
		} else if measure.curStoreOp != StoreOpInvalid {
			measure.storeAccumDur[byte(measure.curStoreOp)] += elapsed
		} else if measure.curNativeOp != NativeOpInvalid {
			measure.nativeAccumDur[byte(measure.curNativeOp)] += elapsed
		}
	}
	measure.curStart = now
}

// SwitchOpCode finalizes the current op's elapsed time and
// starts timing a new CPU op. Returns the old op code so the
// caller can pass it to ResumeOpCode when done.
func SwitchOpCode(code CPUOp) CPUOp {
	if code == CPUOpInvalid {
		panic("the OpCode is invalid")
	}
	old := measure.curCPUOp
	reapTime()
	measure.curCPUOp = code
	measure.curStoreOp = StoreOpInvalid
	measure.curNativeOp = NativeOpInvalid
	measure.opCounts[byte(code)]++
	return old
}

// ResumeOpCode resumes a previous CPU op without incrementing its count.
func ResumeOpCode(code CPUOp) {
	reapTime()
	measure.curCPUOp = code
	measure.curStoreOp = StoreOpInvalid
	measure.curNativeOp = NativeOpInvalid
}

// StopOpCode finalizes the current op. Used at OpHalt/return.
func StopOpCode() {
	reapTime()
	measure.curCPUOp = CPUOpInvalid
	measure.curStoreOp = StoreOpInvalid
	measure.curNativeOp = NativeOpInvalid
	measure.curStart = measure.timeZero
}

// ---- Store operations ----

// StartStore suspends the current op timer and begins a
// store operation. Returns the old (CPUOp, StoreOp) so the
// caller can pass them to StopStore to restore.
func StartStore(storeCode StoreOp) (CPUOp, StoreOp) {
	oldCPU := measure.curCPUOp
	oldStore := measure.curStoreOp
	reapTime()
	measure.curCPUOp = CPUOpInvalid
	measure.curStoreOp = storeCode
	measure.curNativeOp = NativeOpInvalid
	measure.storeCounts[byte(storeCode)]++
	return oldCPU, oldStore
}

// StopStore ends the store operation, records its size,
// then resumes the previous op.
func StopStore(storeCode StoreOp, oldCPU CPUOp, oldStore StoreOp, size int) {
	reapTime()
	measure.storeAccumSize[byte(storeCode)] += int64(size)
	measure.curCPUOp = oldCPU
	measure.curStoreOp = oldStore
	measure.curNativeOp = NativeOpInvalid
}

// ---- Native operations ----

func StartNative(nativeCode NativeOp) CPUOp {
	if nativeCode == NativeOpInvalid {
		panic("the NativeCode is invalid")
	}
	old := measure.curCPUOp
	reapTime()
	measure.curStart = measure.timeZero // discard; GC follows
	runtime.GC()
	measure.curStart = time.Now() // fresh timestamp after GC
	measure.curCPUOp = CPUOpInvalid
	measure.curStoreOp = StoreOpInvalid
	measure.curNativeOp = nativeCode
	measure.nativeCounts[byte(nativeCode)]++
	return old
}

func StopNative(nativeCode NativeOp, old CPUOp) {
	reapTime()
	measure.curCPUOp = old
	measure.curStoreOp = StoreOpInvalid
	measure.curNativeOp = NativeOpInvalid
}
