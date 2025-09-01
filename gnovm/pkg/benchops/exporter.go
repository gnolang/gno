package benchops

import (
	"encoding/binary"
	"log"
	"math"
	"os"
	"time"
)

// the byte size of a exported record
const RecordSize int = 11

var fileWriter *exporter

func initExporter(fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		panic("could not create benchmark file: " + err.Error())
	}

	fileWriter = &exporter{
		file: file,
	}
}

type exporter struct {
	file *os.File
}

// export code, duration, size in a 10 bytes record
// byte 1: OpCode
// byte 2: StoreCode
// byte 3: NativeCode
// byte 4-7: Duration
// byte 8-11: Size
func (e *exporter) export(code Code, elapsedTime time.Duration, size int64) {
	// the MaxUint32 is 4294967295. It represents 4.29 seconds in duration or 4G bytes.
	// It panics not only for overflow protection, but also for abnormal measurements.
	if elapsedTime > math.MaxUint32 {
		log.Fatalf("elapsedTime %d out of uint32 range", elapsedTime)
	}
	if size > math.MaxUint32 {
		log.Fatalf("size %d out of uint32 range", size)
	}

	buf := []byte{code[0], code[1], code[2], 0, 0, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(buf[3:], uint32(elapsedTime))
	binary.LittleEndian.PutUint32(buf[7:], uint32(size))
	_, err := e.file.Write(buf)
	if err != nil {
		panic("could not write to benchmark file: " + err.Error())
	}
}

func (e *exporter) close() {
	e.file.Sync()
	e.file.Close()
}

func FinishStore() {
	for i := range 256 {
		count := measure.storeCounts[i]

		if count == 0 {
			continue
		}
		// check unstopped timer
		if measure.storeStartTime[i] != measure.timeZero {
			panic("timer should have stopped before FinishRun")
		}

		code := [3]byte{0x00, byte(i), 0x00}

		fileWriter.export(
			code,
			measure.storeAccumDur[i]/time.Duration(count),
			measure.storeAccumSize[i]/count,
		)
	}
}

func FinishRun() {
	for i := range 256 {
		if measure.opCounts[i] == 0 {
			continue
		}
		// check unstopped timer
		if measure.opStartTime[i] != measure.timeZero {
			panic("timer should have stopped before FinishRun")
		}

		code := [3]byte{byte(i), 0x00, 0x00}
		fileWriter.export(code, measure.opAccumDur[i]/time.Duration(measure.opCounts[i]), 0)
	}
	ResetRun()
}

func FinishNative() {
	for i := range 256 {
		count := measure.nativeCounts[i]

		if count == 0 {
			continue
		}
		// check unstopped timer
		if measure.nativeStartTime[i] != measure.timeZero {
			panic("timer should have stopped before FinishRun")
		}

		code := [3]byte{0x00, 0x00, byte(i)}

		fileWriter.export(
			code,
			measure.nativeAccumDur[i]/time.Duration(count),
			0,
		)
	}
}

// It reset each machine Runs
func ResetRun() {
	measure.opCounts = [256]int64{}
	measure.opAccumDur = [256]time.Duration{}
	measure.opStartTime = [256]time.Time{}
	measure.curOpCode = invalidCode
	measure.isOpCodeStarted = false
}

func Finish() {
	fileWriter.close()
}
