package fuzzinglib

type StateMachine struct {
	Input_count          uint
	All_lines            Coverage
	Covered_lines        Coverage
	Path_hash2Priority   []uint // 배열. 가장 간단.
	Path_hash2Seed       []Seed
	Priority2Path_hash   []Uint_LinkedList // 링크드 리스트의 배열
	Path_hash2Trio_queue []Uint_Trio_Queue
}

// New_StateMachine: 새로운 StateMachine 생성
func New_StateMachine() *StateMachine {
	return &StateMachine{
		Input_count:          0,                          // 기본값으로 0 설정
		Path_hash2Priority:   make([]uint, 0),            // 빈 uint 배열
		Path_hash2Seed:       make([]Seed, 0),            // 빈 Seed 배열
		Priority2Path_hash:   make([]Uint_LinkedList, 0), // 빈 링크드 리스트 배열
		Path_hash2Trio_queue: make([]Uint_Trio_Queue, 0), // 빈 Trio Queue 배열
	}
}

// func main() {
// 	state_machine := New_StateMachine()
// 	println(state_machine)
// }

// 기반 상태, 기반 자료구조 생성.
// 그것 기반으로 동작 설정.
// 또한 그거 기반 메서드 설정. 자세한건
// github의 fuzz참고.
