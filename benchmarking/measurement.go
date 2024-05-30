package benchmarking

import (
	"time"
)

type measurement struct {
	*timer
	code Code
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

func (m *measurement) end(size int) {
	m.stop()
	fileWriter.export(m.code, m.elapsedTime, size)
}
