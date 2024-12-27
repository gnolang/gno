package fuzzinglib

import (
	"strconv"
	"strings"
	"testing"
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
	Pid           uint
	Id            uint
	Gen           uint
	IsCoordinated bool
	HashNumber    HashNumber
	Content       []interface{}
	Result        interface{}
}

const MAX_CAPACITY uint64 = 50000

type StateMachine struct {
	Trials      uint
	Input_count uint

	All_Coverage     Coverage // []coveredline
	Covered_Coverage Coverage // []coveredline

	// Path_Hash2Coverage *PathHash2Coverage_table
	InitSeeds []Seed
	SeedType  []SupportedType

	Priority_Cache         uint
	HashNumber2Seeds       *[MAX_CAPACITY]*Seed_Queue             // 시드 큐의 배열 포인터
	HashNumber2Priority    *[MAX_CAPACITY]uint                    // 배열. 가장 간단.
	Priority2HashNumberULL *[MAX_CAPACITY]*Unique_Uint_LinkedList // 시간복잡도 줄인 특수

	Crash_Logger *Crash_Logger
}

// New_StateMachine: 새로운 StateMachine 생성
func New_StateMachine() *StateMachine {
	setted_AllULL := func() *[MAX_CAPACITY]*Unique_Uint_LinkedList {
		arr := &[MAX_CAPACITY]*Unique_Uint_LinkedList{}
		for i := range arr {
			arr[i] = New_Unique_Uint_LinkedList()
		}
		return arr
	}()
	return &StateMachine{
		Input_count: 0,
		// TODO: cmd로 반복 가능 횟수 설정정하게 하기
		// 지금은 3만회로 설정
		Trials: 10000,
		// 기본값으로 0 설정
		All_Coverage:     Coverage{{co_name: "nil", co_line: 0}},
		Covered_Coverage: Coverage{{co_name: "nil", co_line: 0}},

		Priority_Cache: 1,
		// Path_Hash2Coverage: New_PathHash2Coverage_table(),

		HashNumber2Priority:    &[MAX_CAPACITY]uint{},        // 빈 uint 배열
		HashNumber2Seeds:       &[MAX_CAPACITY]*Seed_Queue{}, // 빈 Seed 배열
		Priority2HashNumberULL: setted_AllULL,                // 빈 링크드 리스트 배열

		Crash_Logger: New_Crash_Logger(),
	}
}

func (sm *StateMachine) CoordinateMachine(seed Seed) (is_end bool) {
	// 이미 코디네이팅 되었으면 큐 삽입 후 바로 리턴 (우선순위 조정 x)
	if seed.IsCoordinated {
		sm.HashNumber2Seeds[seed.HashNumber].Enqueue(seed)
		return false
	}
	hn := seed.HashNumber
	seed.IsCoordinated = true
	sm.Input_count++
	if sm.Input_count%1000 == 0 {
		println(sm.Input_count, "회 진행")
	}
	// 미등록 해시 넘버시 새 큐 매칭
	if sm.HashNumber2Seeds[hn] == nil {
		sm.HashNumber2Seeds[hn] = New_Seed_Queue()
	}
	sm.HashNumber2Seeds[hn].Enqueue(seed)
	old_priority := sm.HashNumber2Priority[hn]
	sm.HashNumber2Priority[hn]++
	// 얻데이트 우선순위
	updated_priority := sm.HashNumber2Priority[hn]

	// 처음 들어온 해시넘버일 경우
	if updated_priority == 1 {
		// 첫 해시넘버 시 우선순위 캐싱 초기화
		sm.Priority_Cache = 1
		sm.Priority2HashNumberULL[updated_priority].Append(uint(hn))
	} else {
		sm.Priority2HashNumberULL[old_priority].Delete(uint(hn))
		sm.Priority2HashNumberULL[updated_priority].Append(uint(hn))
	}
	if sm.Input_count >= sm.Trials {
		is_end = true
		return is_end
	} else {
		is_end = false
		return is_end
	}
}

func (sm *StateMachine) PopSeeds() []Seed {
	var highest_hn uint
	for {
		// 가장 높은 우선순위의 hashnumber를 선택
		hn, isExist := sm.Priority2HashNumberULL[sm.Priority_Cache].Peek()
		if !isExist {
			sm.Priority_Cache++
			continue
		}
		highest_hn = hn
		break
	}

	popedSeed_1, isOnce := sm.HashNumber2Seeds[highest_hn].Dequeue()
	if !isOnce {
		panic("logical internal error: it must has more than one seed")
	}
	peekedSeed_2, err := sm.HashNumber2Seeds[highest_hn].Peek()
	if err != nil {
		return []Seed{popedSeed_1}
	}
	pid_1 := popedSeed_1.Pid
	pid_2 := peekedSeed_2.Pid
	if pid_1 == pid_2 {
		popedSeed_2, _ := sm.HashNumber2Seeds[highest_hn].Dequeue()
		return []Seed{popedSeed_1, popedSeed_2}
	} else {
		return []Seed{popedSeed_1}
	}
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
	sb.WriteString(sliceToString(sm.HashNumber2Priority[:]))
	sb.WriteString("\n")

	sb.WriteString("  나머진 추가 구현해보삼: ")

	// 반환
	return sb.String()
}
