package benchmarking

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
	timeZero        time.Time

	storeCounts    [256]int64
	storeAccumDur  [256]time.Duration
	storeAccumSize [256]int64
	storeStartTime [256]time.Time
	curStoreCode   byte
}

func InitMeasure() {
	measure = bench{
		// this will be called to reset each benchmarking
		isOpCodeStarted: false,
		curOpCode:       invalidCode,
		curStoreCode:    invalidCode,
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

// StopMeasurement ends the current measurement and resumes the previous one
// if one exists. It accepts the number of bytes that were read/written to/from
// the store. This value is zero if the operation is not a read or write.
func StopOpCode() {
	code := measure.curOpCode
	if measure.opStartTime[code] == measure.timeZero {
		panic("Can not stop a stopped timer")
	}
	measure.opAccumDur[code] += time.Since(measure.opStartTime[code])
	measure.opStartTime[code] = measure.timeZero // stop the timer
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
		panic("Can not stop a stopped timer")
	}

	measure.storeAccumDur[code] += time.Since(measure.storeStartTime[code])
	measure.storeStartTime[code] = measure.timeZero // stop the timer
	measure.storeAccumSize[code] += int64(size)
	measure.curStoreCode = invalidCode
}
