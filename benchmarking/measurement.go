package benchmarking

import (
	"time"
)

type measurement struct {
	*timer
	code       Code
	allocation uint32
}

func startNewMeasurement(code Code) *measurement {
	return &measurement{
		timer: &timer{startTime: time.Now()},
		code:  code,
	}
}

func (m *measurement) pause() {
	m.stop()
}

func (m *measurement) resume() {
	m.start()
}

func (m *measurement) end(size uint32) {
	m.stop()
	if size != 0 && m.allocation != 0 {
		panic("measurement cannot have both allocation and size")
	} else if size == 0 {
		size = m.allocation
	}

	fileWriter.export(m.code, m.elapsedTime, size)
}
