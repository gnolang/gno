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
	opCounts        [256]int64
	opAccumDur      [256]time.Duration
	opStartTime     [256]time.Time
	isOpCodeStarted bool
	curOpCode       byte
	timeZero        time.Time

	storeCounts    [256]int64
	storeAccumDur  [256]time.Duration
	storeAccumSize [256]int64
	storeStartTime [256]time.Time
	curStoreCode   byte

	nativeCounts    [256]int64
	nativeAccumDur  [256]time.Duration
	nativeStartTime [256]time.Time
	isNativeStarted bool
	curNativeCode   byte
	timeZeroNative  time.Time
}

func InitMeasure() {
	measure = bench{
		// this will be called to reset each benchmarking
		isOpCodeStarted: false,
		curOpCode:       invalidCode,
		curStoreCode:    invalidCode,
		curNativeCode:   invalidCode,
	}
}

func StartOpCode(code byte) {
	if code == invalidCode {
		panic("the OpCode is invalid")
	}
	if !measure.opStartTime[code].Equal(measure.timeZero) {
		panic("Can not start a non-stopped timer")
	}
	measure.opStartTime[code] = time.Now()
	measure.opCounts[code]++

	measure.isOpCodeStarted = true
	measure.curOpCode = code
}

// Stop the current measurement
func StopOpCode() {
	code := measure.curOpCode
	if measure.opStartTime[code].Equal(measure.timeZero) {
		panic("Can not stop a stopped timer")
	}
	measure.opAccumDur[code] += time.Since(measure.opStartTime[code])
	measure.opStartTime[code] = measure.timeZero // stop the timer
	measure.isOpCodeStarted = false
}

// Pause current opcode measurement
func PauseOpCode() {
	if !measure.isOpCodeStarted {
		return
	}
	if measure.curOpCode == invalidCode {
		panic("Can not Pause timer of an invalid OpCode")
	}
	code := measure.curOpCode
	if measure.opStartTime[code].Equal(measure.timeZero) {
		panic("Should not pause a stopped timer")
	}
	measure.opAccumDur[code] += time.Since(measure.opStartTime[code])
	measure.opStartTime[code] = measure.timeZero
}

// Resume resumes current measurement
func ResumeOpCode() {
	if !measure.isOpCodeStarted {
		return
	}
	if measure.curOpCode == invalidCode {
		panic("Can not resume timer of an invalid OpCode")
	}

	code := measure.curOpCode

	if measure.opStartTime[code] != measure.timeZero {
		panic("Should not resume a running timer")
	}
	measure.opStartTime[code] = time.Now()
}

func StartStore(code byte) {
	if measure.storeStartTime[code] != measure.timeZero {
		panic("Can not start a non-stopped timer")
	}
	measure.storeStartTime[code] = time.Now()
	measure.storeCounts[code]++
	measure.curStoreCode = code
}

// assume there is no recursive call for store.
func StopStore(size int) {
	code := measure.curStoreCode

	if measure.storeStartTime[code].Equal(measure.timeZero) {
		panic("Can not stop a stopped timer")
	}

	measure.storeAccumDur[code] += time.Since(measure.storeStartTime[code])
	measure.storeStartTime[code] = measure.timeZero // stop the timer
	measure.storeAccumSize[code] += int64(size)
	measure.curStoreCode = invalidCode
}

func StartNative(code byte) {
	if code == invalidCode {
		panic("the OpCode is invalid")
	}
	if !measure.nativeStartTime[code].Equal(measure.timeZero) {
		panic("Can not start a non-stopped timer")
	}
	runtime.GC() // run GC before starting native code timing
	measure.nativeStartTime[code] = time.Now()
	measure.nativeCounts[code]++

	measure.isNativeStarted = true
	measure.curNativeCode = code
}

func StopNative() {
	if !measure.isNativeStarted {
		return
	}
	if measure.curNativeCode == invalidCode {
		panic("Can not stop timer of an invalid OpCode")
	}

	code := measure.curNativeCode

	if measure.nativeStartTime[code].Equal(measure.timeZero) {
		panic("Can not stop a stopped timer")
	}

	measure.nativeAccumDur[code] += time.Since(measure.nativeStartTime[code])
	measure.nativeStartTime[code] = measure.timeZero // stop the timer
	measure.curNativeCode = invalidCode
}
