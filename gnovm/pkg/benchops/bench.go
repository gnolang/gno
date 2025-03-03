package benchops

import (
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

	storeCounts    [256]int64
	storeAccumDur  [256]time.Duration
	storeAccumSize [256]int64
	storeStartTime [256]time.Time
	curStoreCode   byte

	gcCounts        [2]int64
	gcAccumDur      [2]time.Duration
	gcStartTime     [2]time.Time
	isGCCodeStarted bool
	curGCCode       byte

	timeZero time.Time
}

func InitMeasure() {
	measure = bench{
		// this will be called to reset each benchmarking
		isOpCodeStarted: false,
		isGCCodeStarted: false,
		curOpCode:       invalidCode,
		curStoreCode:    invalidCode,
		curGCCode:       invalidCode,
	}
}

func StartOpCode(code byte) {
	if code == invalidCode {
		panic("the OpCode is invalid")
	}
	if measure.opStartTime[code] != measure.timeZero {
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
	if measure.opStartTime[code] == measure.timeZero {
		panic("Can not stop a stopped timer for OpCode")
	}
	measure.opAccumDur[code] += time.Since(measure.opStartTime[code])
	measure.opStartTime[code] = measure.timeZero // stop the timer
	measure.isOpCodeStarted = false
}

// Pause current opcode measurement
func PauseOpCode() {
	if measure.isOpCodeStarted == false {
		return
	}
	if measure.curOpCode == invalidCode {
		panic("Can not Pause timer of an invalid OpCode")
	}
	code := measure.curOpCode
	if measure.opStartTime[code] == measure.timeZero {
		panic("Should not pause a stopped timer")
	}
	measure.opAccumDur[code] += time.Since(measure.opStartTime[code])
	measure.opStartTime[code] = measure.timeZero
}

// Resume resumes current measurement
func ResumeOpCode() {
	if measure.isOpCodeStarted == false {
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

	if measure.storeStartTime[code] == measure.timeZero {
		panic("Can not stop a stopped timer for store")
	}

	measure.storeAccumDur[code] += time.Since(measure.storeStartTime[code])
	measure.storeStartTime[code] = measure.timeZero // stop the timer
	measure.storeAccumSize[code] += int64(size)
	measure.curStoreCode = invalidCode
}

func StartGCCode(code byte) {
	if code == invalidCode {
		panic("the GCCode is invalid")
	}
	if measure.gcStartTime[code] != measure.timeZero {
		panic("Can not start a non-stopped timer")
	}

	measure.gcStartTime[code] = time.Now()
	measure.gcCounts[code]++

	measure.isGCCodeStarted = true
	measure.curGCCode = code
}

// Pause current gc measurement
func PauseGCCode() {
	if measure.isGCCodeStarted == false {
		return
	}
	if measure.curGCCode == invalidCode {
		panic("Can not Pause timer of an invalid GCCode")
	}
	code := measure.curGCCode
	if measure.gcStartTime[code] == measure.timeZero {
		panic("Should not pause a stopped timer")
	}
	measure.gcAccumDur[code] += time.Since(measure.gcStartTime[code])
	measure.gcStartTime[code] = measure.timeZero
}

// Resume resumes current gc measurement
func ResumeGCCode() {
	if measure.isGCCodeStarted == false {
		return
	}
	if measure.curGCCode == invalidCode {
		panic("Can not resume timer of an invalid GCCode")
	}

	code := measure.curGCCode

	if measure.gcStartTime[code] != measure.timeZero {
		panic("Should not resume a running timer")
	}
	measure.gcStartTime[code] = time.Now()
}

// Stop the current measurement
func StopGCCode() {
	code := measure.curGCCode
	if measure.gcStartTime[code] == measure.timeZero {
		panic("Can not stop a stopped timer for GC")
	}

	measure.gcAccumDur[code] += time.Since(measure.gcStartTime[code])
	measure.gcStartTime[code] = measure.timeZero // stop the timer
	measure.isGCCodeStarted = false
}

func IsGCMeasureStarted() bool {
	return measure.isGCCodeStarted
}
