package fuzzinglib

import "errors"

// Seed_Stack 구조체 정의
type Seed_Stack struct {
	data []Seed
}

// NewSeed_Stack: 새로운 Seed_Stack 생성
func New_Seed_Stack() *Seed_Stack {
	return &Seed_Stack{
		data: make([]Seed, 0), // 빈 슬라이스로 초기화
	}
}

// Push: 스택에 Seed 추가
func (s *Seed_Stack) Push(seed Seed) {
	s.data = append(s.data, seed)
}

// Pop: 스택에서 마지막 Seed 제거 및 반환
func (s *Seed_Stack) Pop() (Seed, bool) {
	if len(s.data) == 0 {
		return Seed{}, false // 빈 스택인 경우
	}
	// 마지막 원소를 가져옴
	top := s.data[len(s.data)-1]
	// 슬라이스에서 제거
	s.data = s.data[:len(s.data)-1]
	return top, true
}

// Peek: 스택의 마지막 Seed 조회
func (s *Seed_Stack) Peek() (Seed, error) {
	if len(s.data) == 0 {
		return Seed{}, errors.New("stack is empty")
	}
	return s.data[len(s.data)-1], nil
}

// IsEmpty: 스택이 비어있는지 확인
func (s *Seed_Stack) IsEmpty() bool {
	return len(s.data) == 0
}

// Size: 스택의 크기 반환
func (s *Seed_Stack) Size() int {
	return len(s.data)
}

// func main() {
// 	// 새로운 Seed_Stack 생성
// 	stack := NewSeed_Stack()

// 	// Seed 추가
// 	stack.Push(Seed{Gen: 1, IsUsed: false, Content: "First Seed"})
// 	stack.Push(Seed{Gen: 2, IsUsed: true, Content: "Second Seed"})
// 	stack.Push(Seed{Gen: 3, IsUsed: false, Content: "Third Seed"})

// 	// 스택 상태 출력
// 	println("Stack Size:", stack.Size())

// 	// Peek 호출
// 	top, err := stack.Peek()
// 	if err != nil {
// 		println("Error:", err.Error())
// 	} else {
// 		println("Top Element:", top.Content)
// 	}

// 	// 스택에서 원소 제거
// 	for !stack.IsEmpty() {
// 		seed, _ := stack.Pop()
// 		println("Popped Seed:", seed.Content)
// 	}

// 	// 스택이 비었는지 확인
// 	println("Is Stack Empty?", stack.IsEmpty())
// }
