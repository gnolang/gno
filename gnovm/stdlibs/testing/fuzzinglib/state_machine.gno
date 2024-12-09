package fuzzinglib

import "strings"

type StateMachine struct {
	Input_count uint

	All_Coverage     Coverage // []uint
	Covered_Coverage Coverage // []uint

	Path_Hash2Coverage *PathHash2Coverage_table
	First_seed         Seed

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
		All_Coverage:     Coverage{0},
		Covered_Coverage: Coverage{0},

		Path_Hash2Coverage: New_PathHash2Coverage_table(),

		Path_hash2Priority:   make([]uint, 0),                           // 빈 uint 배열
		Path_hash2Seeds:      []*Seed_Stack{New_Seed_Stack()},           // 빈 Seed 배열
		Priority2Path_hash:   []*Uint_LinkedList{New_Uint_LinkedList()}, // 빈 링크드 리스트 배열
		Path_hash2Trio_queue: []*Uint_Trio_Queue{New_Uint_Trio_Queue()}, // 빈 Trio Queue 배열

		Crash_Logger: New_Crash_Logger(),
	}
}

// 최초의 입력을 받음.
// 함수와 시드에 대한 입력.
func (sm *StateMachine) Format_StateMachine(f interface{}, s string) {
	first_seed := Seed{
		Gen:     1,
		IsUsed:  false,
		Content: s,
		Result:  nil,
	}
	sm.First_seed = first_seed
	// 원래는 함수를 입력받습니다.
	// 그러나 현재 커버리지 기능은 개발을 기다려야 하는 상황입니다.
	// 때문에 가상의 함수, "SomeFunc"에 대해 완전히 작동하는 커버리지 함수 만들었습니다.
	sm.All_Coverage = Get_AllCoverage()
	result := Get_PseudoCoverage_of_SomeFunc(s)
	if sm.Crash_Logger.IsCrash(result) {
		sm.Crash_Logger.AddCase(first_seed, result)
	}
	// TODO: 추가로 마저 스테이트 수정!! 배열 파트!!
	// TODO: 라인 추가 등
	// 힌트는, 밑에 포멧팅 출력해 보면서, 노션 보변서 감 잡으삼
	// result := Get_PseudoCoverage_of_SomeFunc(s)
}

func main() {
	state_machine := New_StateMachine()
	state_machine.Format_StateMachine('some_func', "apple")
	println(state_machine.ToString())
}

// print_machine: StateMachine을 예쁘게 출력하는 메서드
func (sm *StateMachine) ToString() string {
	// 문자열 빌더 사용
	var sb strings.Builder

	sb.WriteString("StateMachine:\n")
	sb.WriteString("  Input_count: ")
	sb.WriteString(uintToString(sm.Input_count))
	sb.WriteString("\n")

	sb.WriteString("  All_lines: ")
	sb.WriteString(sliceToString(sm.All_Coverage))
	sb.WriteString("\n")

	sb.WriteString("  Covered_lines: ")
	sb.WriteString(sliceToString(sm.Covered_Coverage))
	sb.WriteString("\n")

	sb.WriteString("  First_seed:\n")
	sb.WriteString("    Gen: ")
	sb.WriteString(uintToString(sm.First_seed.Gen))
	sb.WriteString("\n")
	sb.WriteString("    IsUsed: ")
	sb.WriteString(boolToString(sm.First_seed.IsUsed))
	sb.WriteString("\n")
	sb.WriteString("    Content: ")
	sb.WriteString(sm.First_seed.Content)
	sb.WriteString("\n")
	sb.WriteString("    Result: ")
	sb.WriteString(interfaceToString(sm.First_seed.Result))
	sb.WriteString("\n")

	sb.WriteString("  Path_hash2Priority: ")
	sb.WriteString(sliceToString(sm.Path_hash2Priority))
	sb.WriteString("\n")

	sb.WriteString("  나머진 추가 구현해보삼: ")

	// 반환
	return sb.String()
}
