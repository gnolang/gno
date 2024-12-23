package fuzzinglib

import (
	"strconv"
	"strings"
)

type SupportedType string

const (
	Byte_Array SupportedType = "[]byte"
	String     SupportedType = "string"
	Bool       SupportedType = "bool"
	Byte       SupportedType = "byte"
	Rune       SupportedType = "rune"
	Float32    SupportedType = "float32"
	Float64    SupportedType = "float64"
	Int        SupportedType = "int"
	Int8       SupportedType = "int8"
	Int16      SupportedType = "int16"
	Int32      SupportedType = "int32"
	Int64      SupportedType = "int64"
	Uint       SupportedType = "uint"
	Uint8      SupportedType = "uint8"
	Uint16     SupportedType = "uint16"
	Uint32     SupportedType = "uint32"
	Uint64     SupportedType = "uint64"
)

type Seed struct {
	Gen     uint
	IsUsed  bool
	Content []interface{}
	Result  interface{}
}
type StateMachine struct {
	Input_count uint

	All_Coverage     Coverage // []coveredline
	Covered_Coverage Coverage // []coveredline

	// Path_Hash2Coverage *PathHash2Coverage_table
	InitSeeds []Seed
	SeedType  []SupportedType

	Path_hash2Priority   []uint             // 배열. 가장 간단.
	Path_hash2Seeds      []*Seed_Stack      // 시드 스택의 배열 포인터
	Priority2Path_hash   []*Uint_LinkedList // 링크드 리스트의 배열 포인터
	Path_hash2Trio_queue []*Uint_Trio_Queue // 큐의 배열 포인터

	Crash_Logger *Crash_Logger
}

// New_StateMachine: 새로운 StateMachine 생성
func New_StateMachine() *StateMachine {
	return &StateMachine{
		Input_count: 0,
		// 기본값으로 0 설정
		All_Coverage:     Coverage{{co_name: "nil", co_line: 0}},
		Covered_Coverage: Coverage{{co_name: "nil", co_line: 0}},

		// Path_Hash2Coverage: New_PathHash2Coverage_table(),

		Path_hash2Priority:   make([]uint, 0),                           // 빈 uint 배열
		Path_hash2Seeds:      []*Seed_Stack{New_Seed_Stack()},           // 빈 Seed 배열
		Priority2Path_hash:   []*Uint_LinkedList{New_Uint_LinkedList()}, // 빈 링크드 리스트 배열
		Path_hash2Trio_queue: []*Uint_Trio_Queue{New_Uint_Trio_Queue()}, // 빈 Trio Queue 배열

		Crash_Logger: New_Crash_Logger(),
	}
}

// 최초의 입력을 받음.
// 함수와 시드에 대한 입력.
func (sm *StateMachine) Format_StateMachine(f interface{}) {
	sm.All_Coverage = Get_AllCoverage()

	for _, e := range sm.InitSeeds {
		// TODO: 쩔 수 없음. 현재는  string에 대한 하드코딩밖에 안됨
		// seed가 하나뿐
		// 심지어 다변수는 지원도 안함
		// 사후처리도 불가.
		str, ok := e.Content[0].(string)
		if !ok {
			panic("nooooot a string")
		}
		result := Get_Coverage_of_target_func(str)
		println("result", result)
	}

	// if sm.Crash_Logger.IsCrash(result) {
	// 	sm.Crash_Logger.AddCase(first_seed, result)
	// }
	// TODO: 추가로 마저 스테이트 수정!! 배열 파트!!
	// TODO: 라인 추가 등, 그리고 그를 위한 기반 함수들도 디자인.
	// 힌트는, 밑에 포멧팅 출력해 보면서, 노션 보변서 감 잡으삼
	// result := Get_PseudoCoverage_of_SomeFunc(s)
}

// func main() {
// 	state_machine := New_StateMachine()
// 	state_machine.Format_StateMachine("some_func", "apple")
// 	println(state_machine.ToString())
// }

// print_machine: StateMachine을 예쁘게 출력하는 메서드
func (sm *StateMachine) ToString() string {
	// 문자열 빌더 사용
	var sb strings.Builder

	sb.WriteString("StateMachine:\n")
	sb.WriteString("  Input_count: ")
	sb.WriteString(uintToString(sm.Input_count))
	sb.WriteString("\n")

	sb.WriteString("  All_lines: ")
	sb.WriteString(coverageToString(sm.All_Coverage)) // 수정된 부분
	sb.WriteString("\n")

	sb.WriteString("  Covered_lines: ")
	sb.WriteString(coverageToString(sm.Covered_Coverage)) // 수정된 부분
	sb.WriteString("\n")

	sb.WriteString("  Path_hash2Priority: ")
	sb.WriteString(sliceToString(sm.Path_hash2Priority))
	sb.WriteString("\n")

	sb.WriteString("  나머진 추가 구현해보삼: ")

	// 반환
	return sb.String()
}
