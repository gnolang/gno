package benchmarking

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

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

func (e *exporter) export(code Code, elapsedTime time.Duration, size uint32) {
	buf := []byte{code[0], code[1], 0, 0, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(buf[2:], uint32(elapsedTime))
	binary.LittleEndian.PutUint32(buf[6:], size)
	_, err := e.file.Write(buf)
	if err != nil {
		panic("could not write to benchmark file: " + err.Error())
	}
}

func (e *exporter) close() {
	e.file.Sync()
	e.file.Close()
}

func Finish() {
	fmt.Println("## StackSize: ", stackSize)

	fileWriter.close()
}
