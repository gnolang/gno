package benchops

import (
	"encoding/binary"
	"log"
	"math"
	"os"
	"time"
)

// the byte size of an exported record
const RecordSize int = 14

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

// export code, duration, size, count in a 14 byte record
// byte 1:    Type (0x01=OpCode, 0x02=StoreCode, 0x03=NativeCode)
// byte 2:    OpCode, StoreCode, or NativeCode
// bytes 3-6: Total duration (uint32, nanoseconds)
// bytes 7-10: Total size (uint32, bytes; 0 for opcodes/native)
// bytes 11-14: Count (uint32)
func (e *exporter) export(code Code, totalTime time.Duration, totalSize int64, count int64) {
	// the MaxUint32 is 4294967295. It represents 4.29 seconds in duration or 4G bytes.
	// It panics not only for overflow protection, but also for abnormal measurements.
	if totalTime > math.MaxUint32 {
		log.Fatalf("totalTime %d out of uint32 range", totalTime)
	}
	if totalSize > math.MaxUint32 {
		log.Fatalf("totalSize %d out of uint32 range", totalSize)
	}
	if count > math.MaxUint32 {
		log.Fatalf("count %d out of uint32 range", count)
	}

	buf := make([]byte, RecordSize)
	buf[0] = code[0]
	buf[1] = code[1]
	binary.LittleEndian.PutUint32(buf[2:], uint32(totalTime))
	binary.LittleEndian.PutUint32(buf[6:], uint32(totalSize))
	binary.LittleEndian.PutUint32(buf[10:], uint32(count))
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

		code := [2]byte{byte(TypeStore), byte(i)}
		fileWriter.export(
			code,
			measure.storeAccumDur[i],
			measure.storeAccumSize[i],
			count,
		)
	}
}

func FinishRun() {
	// Ensure the timeline is stopped.
	if measure.curOpCode != invalidCode {
		StopOpCode()
	}

	for i := range 256 {
		count := measure.opCounts[i]
		if count == 0 {
			continue
		}

		code := [2]byte{byte(TypeOpCode), byte(i)}
		fileWriter.export(code, measure.opAccumDur[i], 0, count)
	}
	ResetRun()
}

func FinishNative() {
	for i := range 256 {
		count := measure.nativeCounts[i]
		if count == 0 {
			continue
		}

		code := [2]byte{byte(TypeNative), byte(i)}
		fileWriter.export(
			code,
			measure.nativeAccumDur[i],
			0,
			count,
		)
	}
}

// ResetRun resets opcode measurements between machine runs.
func ResetRun() {
	measure.opCounts = [256]int64{}
	measure.opAccumDur = [256]time.Duration{}
	measure.curOpCode = invalidCode
	measure.curStart = measure.timeZero
}

func Finish() {
	fileWriter.close()
}
