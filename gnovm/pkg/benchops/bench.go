package benchmarking

import (
	"time"
)

const (
	invalidCode = byte(0x00)
)

var (
	opCounts        = [256]int64{}
	opAccumDur      = [256]time.Duration{}
	opStartTime     = [256]time.Time{}
	isOpCodeStarted = false
	curOpCode       byte
	timeZero        time.Time

	storeCounts    = [256]int64{}
	storeAccumDur  = [256]time.Duration{}
	storeAccumSize = [256]int64{}
	storeStartTime = [256]time.Time{}
	curStoreCode   byte
)

func InitMeasure() {
	// this will be called to reset each benchmarking
	opCounts = [256]int64{}
	opAccumDur = [256]time.Duration{}
	opStartTime = [256]time.Time{}
	isOpCodeStarted = false
	curOpCode = invalidCode

	storeCounts = [256]int64{}
	storeAccumDur = [256]time.Duration{}
	storeAccumSize = [256]int64{}
	storeStartTime = [256]time.Time{}
	curStoreCode = invalidCode
}

func StartOpCode(code byte) {
	if code == invalidCode {
		panic("the OpCode is invalid")
	}
	if opStartTime[code] != timeZero {
		panic("Can not start a non-stopped timer")
	}
	opStartTime[code] = time.Now()
	opCounts[code]++

	isOpCodeStarted = true
	curOpCode = code
}

// StopMeasurement ends the current measurement and resumes the previous one
// if one exists. It accepts the number of bytes that were read/written to/from
// the store. This value is zero if the operation is not a read or write.
func StopOpCode() {
	code := curOpCode
	if opStartTime[code] == timeZero {
		panic("Can not stop a stopped timer")
	}
	opAccumDur[code] += time.Since(opStartTime[code])
	opStartTime[code] = timeZero // stop the timer
}

// Pause current opcode measurement
func PauseOpCode() {
	if isOpCodeStarted == false {
		return
	}
	if curOpCode == invalidCode {
		panic("Can not Pause timer of an invalid OpCode")
	}
	code := curOpCode
	if opStartTime[code] == timeZero {
		panic("Should not pause a stopped timer")
	}
	opAccumDur[code] += time.Since(opStartTime[code])
	opStartTime[code] = timeZero
}

// Resume resumes current measurement
func ResumeOpCode() {
	if isOpCodeStarted == false {
		return
	}
	if curOpCode == invalidCode {
		panic("Can not resume timer of an invalid OpCode")
	}

	code := curOpCode

	if opStartTime[code] != timeZero {
		panic("Should not resume a running timer")
	}
	opStartTime[code] = time.Now()
}

func StartStore(code byte) {
	if storeStartTime[code] != timeZero {
		panic("Can not start a non-stopped timer")
	}
	storeStartTime[code] = time.Now()
	storeCounts[code]++
	curStoreCode = code
}

// assume there is no recursive call for store.
func StopStore(size int) {
	code := curStoreCode

	if storeStartTime[code] == timeZero {
		panic("Can not stop a stopped timer")
	}

	storeAccumDur[code] += time.Since(storeStartTime[code])
	storeStartTime[code] = timeZero // stop the timer

	storeAccumSize[code] += int64(size)

	curStoreCode = invalidCode
}
