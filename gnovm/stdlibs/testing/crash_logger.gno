package testing

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type TestResult struct {
	PanicOccurred bool
	PanicMessage  string
	Error         error
}

type Crash_Case struct {
	HashNumber   uint
	Input        string
	RawContent   []interface{}
	InputCount   uint
	IsPanic      bool
	PanicMessage string
	ErrorMsg     string
	Timestamp    time.Time
}

type Crash_Logger struct {
	data []Crash_Case
}

func New_Crash_Logger() *Crash_Logger {
	return &Crash_Logger{
		data: make([]Crash_Case, 0),
	}
}

func (log *Crash_Logger) AddCase(seed Seed, r TestResult) {
	var crashCase Crash_Case
	if r.Error == nil {
		crashCase = Crash_Case{
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
		crashCase = Crash_Case{
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
	log.data = append(log.data, crashCase)
}

func (log *Crash_Logger) GetCase(index int) (Crash_Case, error) {
	if index < 0 || index >= len(log.data) {
		return Crash_Case{}, errors.New("index out of bounds")
	}
	return log.data[index], nil
}

func (log *Crash_Logger) RemoveCase(index int) error {
	if index < 0 || index >= len(log.data) {
		return errors.New("index out of bounds")
	}
	log.data = append(log.data[:index], log.data[index+1:]...)
	return nil
}

func (log *Crash_Logger) ListCases() []Crash_Case {
	return log.data
}

func (log *Crash_Logger) Size() int {
	return len(log.data)
}

func (log *Crash_Logger) ClearLog() {
	log.data = make([]Crash_Case, 0)
}

type KindInfo struct {
	HashNumber          uint
	RepresentativeError string
	RepresentativeInput string
	IsPanic             bool
	Count               int
}

func (log *Crash_Logger) Kind() []KindInfo {
	lookup := make(map[uint]*KindInfo)

	for _, c := range log.data {
		ki, exists := lookup[c.HashNumber]
		if !exists {

			repMsg := ""

			if c.IsPanic {
				repMsg = c.PanicMessage
			} else {
				repMsg = c.ErrorMsg
			}
			lookup[c.HashNumber] = &KindInfo{
				HashNumber:          c.HashNumber,
				RepresentativeError: repMsg,
				RepresentativeInput: c.Input,
				IsPanic:             c.IsPanic,
				Count:               1,
			}
		} else {
			ki.Count++
		}
	}

	result := make([]KindInfo, 0, len(lookup))
	for _, ki := range lookup {
		result = append(result, *ki)
	}
	return result
}

func (log *Crash_Logger) Samples_of_AllKinds() map[uint][]Crash_Case {
	lookup := make(map[uint][]Crash_Case)
	for _, c := range log.data {
		lookup[c.HashNumber] = append(lookup[c.HashNumber], c)
	}

	for hn, cases := range lookup {
		if len(cases) > 3 {
			cases = cases[:3]
		}
		lookup[hn] = cases
	}
	return lookup
}

type CrashSummary struct {
	TotalCrashes      int
	HashNumbers       []uint
	MostFrequentHN    uint
	MostFrequentCount int
	PanicCount        int
	ErrorCount        int
}

func (log *Crash_Logger) Summary() CrashSummary {
	sm := CrashSummary{}
	sm.TotalCrashes = len(log.data)
	lookup := make(map[uint]int)
	for _, c := range log.data {
		lookup[c.HashNumber]++
		if c.IsPanic {
			sm.PanicCount++
		} else {
			sm.ErrorCount++
		}
	}
	for k := range lookup {
		sm.HashNumbers = append(sm.HashNumbers, k)
	}

	for hn, cnt := range lookup {
		if cnt > sm.MostFrequentCount {
			sm.MostFrequentCount = cnt
			sm.MostFrequentHN = hn
		}
	}
	return sm
}

func PrintSummary(sum CrashSummary) {
	println("----- Crash Log Summary -----")
	println("TotalCrashes:", sum.TotalCrashes)
	println("HashNumbers:", sliceToString(sum.HashNumbers))
	println("MostFrequentHashNumber:", sum.MostFrequentHN, "(occurs", sum.MostFrequentCount, "times)")
	println("PanicCount:", sum.PanicCount)
	println("ErrorCount:", sum.ErrorCount)
	println("--------------------------------")
}

func PrintKinds(kinds []KindInfo) {
	if len(kinds) == 0 {
		println("No crash kinds found.")
		return
	}

	println("----- Crash Kinds -----")
	for _, k := range kinds {
		println("HashNumber:", k.HashNumber,
			", IsPanic:", k.IsPanic,
			", Count:", k.Count,
			", RepresentativeError:", k.RepresentativeError)

		inputStr := k.RepresentativeInput
		println(" Printed Input:", inputStr)
		println(" In machine(Escaped edge bytes) Input:", TransForHuman(inputStr))
		println("-----------------------")
	}
}
