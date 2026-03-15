package benchops

import (
	"runtime"
	"time"
)

const (
	invalidCode = byte(0x00)
)

var measure bench

type bench struct {
	// Opcode timing: single timeline, always one op active.
	opCounts   [256]int64
	opAccumDur [256]time.Duration
	curOpCode  byte
	curStart   time.Time
	timeZero   time.Time

	// Store timing: own accumulators, uses SwitchOpCode for handoff.
	storeCounts    [256]int64
	storeAccumDur  [256]time.Duration
	storeAccumSize [256]int64

	// Native timing.
	nativeCounts   [256]int64
	nativeAccumDur [256]time.Duration
}

func InitMeasure() {
	measure = bench{
		curOpCode: invalidCode,
	}
}

// finalizeCurrent attributes elapsed time since curStart to
// curOpCode and returns the snapshot time.
func finalizeCurrent() time.Time {
	now := time.Now()
	if measure.curOpCode != invalidCode && measure.curStart != measure.timeZero {
		measure.opAccumDur[measure.curOpCode] += now.Sub(measure.curStart)
	}
	measure.curStart = measure.timeZero
	return now
}

// BeginOpCode starts timing the first op in a run loop.
func BeginOpCode(code byte) {
	if code == invalidCode {
		panic("the OpCode is invalid")
	}
	measure.curOpCode = code
	measure.curStart = time.Now()
	measure.opCounts[code]++
}

// SwitchOpCode finalizes the current op's elapsed time and
// starts timing a new op. Returns the old op code so the
// caller can pass it to ResumeOpCode when done.
func SwitchOpCode(code byte) byte {
	if code == invalidCode {
		panic("the OpCode is invalid")
	}
	old := measure.curOpCode
	now := finalizeCurrent()
	measure.curOpCode = code
	measure.curStart = now
	measure.opCounts[code]++
	return old
}

// ResumeOpCode finalizes the current op's elapsed time and
// resumes a previous op without incrementing its count.
func ResumeOpCode(code byte) {
	now := finalizeCurrent()
	measure.curOpCode = code
	measure.curStart = now
}

// StopOpCode finalizes the current op. Used at OpHalt/return.
func StopOpCode() {
	finalizeCurrent()
	measure.curOpCode = invalidCode
}

// ---- Store operations ----

// StartStore suspends the current VM op timer and begins a
// store operation. Returns the old op code for ResumeOpCode.
func StartStore(storeCode byte) byte {
	old := measure.curOpCode
	// Finalize the VM op's time up to now.
	now := finalizeCurrent()
	// Park the timeline — store tracks its own duration.
	measure.curOpCode = invalidCode
	measure.storeCounts[storeCode]++
	// Store the start time in curStart temporarily;
	// StopStore will read it.
	measure.curStart = now
	return old
}

// StopStore ends the store operation, records its duration
// and size, then resumes the previous VM op.
func StopStore(storeCode byte, old byte, size int) {
	now := time.Now()
	if measure.curStart != measure.timeZero {
		measure.storeAccumDur[storeCode] += now.Sub(measure.curStart)
	}
	measure.storeAccumSize[storeCode] += int64(size)
	// Resume the VM op.
	measure.curOpCode = old
	measure.curStart = now
}

// ---- Native operations ----

func StartNative(nativeCode byte) byte {
	if nativeCode == invalidCode {
		panic("the NativeCode is invalid")
	}
	old := measure.curOpCode
	finalizeCurrent() // finalize previous op BEFORE GC
	runtime.GC()
	now := time.Now() // fresh timestamp after GC
	measure.curOpCode = invalidCode
	measure.nativeCounts[nativeCode]++
	measure.curStart = now
	return old
}

func StopNative(nativeCode byte, old byte) {
	now := time.Now()
	if measure.curStart != measure.timeZero {
		measure.nativeAccumDur[nativeCode] += now.Sub(measure.curStart)
	}
	measure.curOpCode = old
	measure.curStart = now
}

// OpAccumDur returns the accumulated duration for an op code.
func OpAccumDur(code byte) time.Duration {
	return measure.opAccumDur[code]
}

// OpCount returns the invocation count for an op code.
func OpCount(code byte) int64 {
	return measure.opCounts[code]
}
