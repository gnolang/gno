package testing

import (
	"errors"
	"strings"
	"time"
)

type TestResult struct {
	PanicOccurred bool
	PanicMessage  string
	Error         error
}

type CrashCase struct {
	HashNumber   uint
	Input        string
	RawContent   []interface{}
	InputCount   uint
	IsPanic      bool
	PanicMessage string
	ErrorMsg     string
	Timestamp    time.Time
}

type CrashLogger struct {
	data []CrashCase
}

func NewCrashLogger() *CrashLogger {
	return &CrashLogger{
		data: make([]CrashCase, 0),
	}
}

func (c *CrashLogger) AddCase(seed Seed, r TestResult) {
	var crashCase CrashCase
	if r.Error == nil {
		crashCase = CrashCase{
			HashNumber:   uint(seed.HashNumber),
			Input:        ContentToString(seed.Content),
			RawContent:   seed.Content,
			InputCount:   seed.Id,
			IsPanic:      r.PanicOccurred,
			PanicMessage: r.PanicMessage,
			ErrorMsg:     "",
			Timestamp:    time.Now(),
		}
	} else {
		crashCase = CrashCase{
			HashNumber:   uint(seed.HashNumber),
			Input:        ContentToString(seed.Content),
			RawContent:   seed.Content,
			InputCount:   seed.Id,
			IsPanic:      r.PanicOccurred,
			PanicMessage: r.PanicMessage,
			ErrorMsg:     strings.TrimSpace(r.Error.Error()),
			Timestamp:    time.Now(),
		}
	}
	c.data = append(c.data, crashCase)
}

func (c *CrashLogger) GetCase(index int) (CrashCase, error) {
	if index < 0 || index >= len(c.data) {
		return CrashCase{}, errors.New("index out of bounds")
	}
	return c.data[index], nil
}

func (c *CrashLogger) RemoveCase(index int) error {
	if index < 0 || index >= len(c.data) {
		return errors.New("index out of bounds")
	}
	c.data = append(c.data[:index], c.data[index+1:]...)
	return nil
}

func (c *CrashLogger) ListCases() []CrashCase {
	return c.data
}

func (c *CrashLogger) Size() int {
	return len(c.data)
}

func (c *CrashLogger) ClearLog() {
	c.data = make([]CrashCase, 0)
}
