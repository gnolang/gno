package fuzzinglib

import (
	"errors"
	"time"
)

type TestResult struct {
	PanicOccurred bool     // panic 여부
	PanicMessage  string   // panic 메시지
	Error         error    // 반환된 에러
	Coverage      Coverage // 반환된 값
}

// Crash_Case 구조체 정의
type Crash_Case struct {
	Input        interface{}
	Coverage     Coverage
	IsPanic      bool
	PanicMessage string
	Error        error
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

func (log *Crash_Logger) IsCrash(r TestResult) bool {
	if r.PanicOccurred {
		return true
	}
	if !(r.Error == nil) {
		return true
	}
	return false
}

// AddCase: 새로운 Crash_Case 추가
func (log *Crash_Logger) AddCase(seed Seed, r TestResult) {
	crashCase := Crash_Case{
		Input:        seed.Content,
		Coverage:     r.Coverage,
		IsPanic:      r.PanicOccurred,
		PanicMessage: r.PanicMessage,
		Error:        r.Error,
		Timestamp:    time.Now(),
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

// func main() {
// 	// Crash_Logger 생성
// 	crashLog := New_Crash_Logger()

// 	// Crash_Case 추가
// 	crashLog.AddCase("input1", Coverage{1, 2, 3}, "SIGSEGV at 0xdeadbeef")
// 	crashLog.AddCase("input2", Coverage{4, 5, 6}, "NullPointerException")

// 	// 크래시 로그 크기 확인
// 	println("Crash Log Size:", crashLog.Size())

// 	// 특정 Crash_Case 조회
// 	case1, err := crashLog.GetCase(0)
// 	if err == nil {
// 		println("First Case - Input:", case1.Input, ", Error:", case1.Error)
// 	}

// 	// 모든 Crash_Case 출력
// 	for i, c := range crashLog.ListCases() {
// 		println("Case", i+1, "- Input:", c.Input, ", Error:", c.Error)
// 	}

// 	// Crash_Case 삭제
// 	if err := crashLog.RemoveCase(0); err == nil {
// 		println("First case removed.")
// 	}

// 	println("After Removed. Size:", crashLog.Size())
// 	// 크래시 로그 초기화
// 	crashLog.ClearLog()
// 	println("Crash Log cleared. Size:", crashLog.Size())
// }
