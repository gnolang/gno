package benchmarking

import (
	"time"
)

const (
	invalidCode    = byte(0x00)
	opStaticTypeOf = byte(0x4A)
)

var (
	opCounts        = [256]int64{}
	opAccumDur      = [256]time.Duration{}
	opStartTime     = [256]time.Time{}
	isOpCodeStarted = false
	curOpCode       byte
	timeZero        time.Time
	stack           = make([]byte, 0, 256)

	storeCounts    = [256]int64{}
	storeAccumDur  = [256]time.Duration{}
	storeAccumSize = [256]int64{}
	storeStartTime = [256]time.Time{}
	curStoreCode   byte
)

func InitStack() {
	// this will be called to reset each benchmarking
	opCounts = [256]int64{}
	opAccumDur = [256]time.Duration{}
	opStartTime = [256]time.Time{}
	isOpCodeStarted = false
	curOpCode = invalidCode
	stack = make([]byte, 0, 256)

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
		// Cornner case: where StopOpCode() OpVoid resumes timer of the code on the stack top
		// the actual OpCode we want to benchmark could be the same as the one resumed at the stack top.
		if code == stack[len(stack)-1] {
			// do nothing
		} else {
			// regular check
			panic("Can not start a non-stopped timer")

		}

	}
	// OpCode, such as OpStaticTypeOf, are pushed on the stack since
	// it invovles recurisve machine.Run()
	if len(stack) > 0 {
		curOpCode = PeekOp()
		PauseOpCode()
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

	if opStartTime[code] == timeZero && code != opStaticTypeOf {
		panic("Can not stop a stopped timer")
	}
	opAccumDur[code] += time.Since(opStartTime[code])
	opStartTime[code] = timeZero // stop the timer

	if len(stack) > 0 {
		curOpCode = PeekOp()
		ResumeOpCode()
	}
	curOpCode = invalidCode

}

// push current op code on stack when an opcode executes recurisve machine.Run()
func PushOp(curCode byte) {

	if curCode == invalidCode {
		panic("Should not put an invalidCode on the stack")
	}
	stack = append(stack, curCode)

}

// peek the top from stack
func PeekOp() byte {

	top := len(stack) - 1
	if top >= 0 {
		return stack[top]
	} else {
		panic("not enough element on the stack")
	}
}

// pop the top from stack and make it current
func PopOp() {

	top := len(stack) - 1
	if top >= 0 {
		code := stack[top]
		stack = stack[:top]
		curOpCode = code
	} else {
		panic("not enough element on the stack")
	}
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

// Resume resumes current measurement on the stack
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
