package testing

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type TestResult struct {
	PanicOccurred bool   // panic 여부
	PanicMessage  string // panic 메시지
	Error         error  // 반환된 에러
}

// Crash_Case 구조체 정의
type Crash_Case struct {
	HashNumber   uint
	Input        string
	RawContent   []interface{}
	InputCount   uint
	IsPanic      bool
	PanicMessage string
	ErrorMsg     string
	Timestamp    time.Time // 충돌 발생 시간 추가
}

// Crash_Log 구조체 정의
type Crash_Logger struct {
	data []Crash_Case
}

// NewCrashLog: 새로운 Crash_Log 생성
func New_Crash_Logger() *Crash_Logger {
	return &Crash_Logger{
		data: make([]Crash_Case, 0),
	}
}

// AddCase: 새로운 Crash_Case 추가
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

// GetCase: 특정 인덱스의 Crash_Case 조회
func (log *Crash_Logger) GetCase(index int) (Crash_Case, error) {
	if index < 0 || index >= len(log.data) {
		return Crash_Case{}, errors.New("index out of bounds")
	}
	return log.data[index], nil
}

// RemoveCase: 특정 인덱스의 Crash_Case 삭제
func (log *Crash_Logger) RemoveCase(index int) error {
	if index < 0 || index >= len(log.data) {
		return errors.New("index out of bounds")
	}
	log.data = append(log.data[:index], log.data[index+1:]...)
	return nil
}

// ListCases: 모든 Crash_Case 출력
func (log *Crash_Logger) ListCases() []Crash_Case {
	return log.data
}

// Size: Crash_Logger의 크기 반환
func (log *Crash_Logger) Size() int {
	return len(log.data)
}

// ClearLog: Crash_Logger 초기화
func (log *Crash_Logger) ClearLog() {
	log.data = make([]Crash_Case, 0)
}

// KindInfo: 각 HashNumber에 대한 간단한 요약 구조
type KindInfo struct {
	HashNumber          uint
	RepresentativeError string // 대표 에러 메시지 or Panic 메시지
	RepresentativeInput string
	IsPanic             bool
	Count               int // 몇 번 발생했는지
}

// kind: Crash_Logger 내 존재하는 고유 HashNumber별 대표 에러/패닉 요약
func (log *Crash_Logger) Kind() []KindInfo {
	// map[HashNumber] -> (대표 메세지, panic여부, count)
	lookup := make(map[uint]*KindInfo)

	for _, c := range log.data {
		ki, exists := lookup[c.HashNumber]
		if !exists {
			// 신규
			repMsg := ""
			// 같은 hn이라 해도 여러 메시지가 있을 수 있다.
			// 같은 라인에서 에러 내도, 에러에 변수 있으면 메시지 다를 수 있기에.
			// 그러나, 여기선 hn의 대표 메시지 하나만 설정함.
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
			// 이미 존재 => Count++
			ki.Count++

			// 대표 메세지는 첫 크래시 케이스 것으로 고정해도 되고,
			// 혹은 새로 들어온 케이스가 Panic인 경우 갱신할 수도 있음. (아래는 단순 Count만 증가)
		}
	}

	// 맵 -> 슬라이스
	result := make([]KindInfo, 0, len(lookup))
	for _, ki := range lookup {
		result = append(result, *ki)
	}
	return result
}

// allSamples: 각 HashNumber별로 "최신순" 3개 Crash_Case를 반환
// return map[HashNumber][]Crash_Case
func (log *Crash_Logger) Samples_of_AllKinds() map[uint][]Crash_Case {
	// 1) HashNumber -> []Crash_Case
	lookup := make(map[uint][]Crash_Case)
	for _, c := range log.data {
		lookup[c.HashNumber] = append(lookup[c.HashNumber], c)
	}

	// 2) 정렬 후, 상위 3개만 추출
	for hn, cases := range lookup {
		// 최신순 정렬(내림차순)
		// 3개 초과면 잘라낸다
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

	// 가장 많이 발생한 HashNumber
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

		// RepresentativeInput도 출력해보자
		inputStr := k.RepresentativeInput
		println(" Printed Input:", inputStr)
		println(" In machine(Escaped edge bytes) Input:", TransForHuman(inputStr))
		println("-----------------------")
	}
}
