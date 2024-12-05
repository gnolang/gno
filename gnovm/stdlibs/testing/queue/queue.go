package main

import "fmt"

// Queue 구조체
type Queue[T any] struct {
	data []T
}

// NewQueue: 새로운 큐 생성
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{data: make([]T, 0)}
}

// Enqueue: 큐에 삽입
func (q *Queue[T]) Enqueue(value T) {
	q.data = append(q.data, value)
}

// Dequeue: 큐에서 제거 후 반환
func (q *Queue[T]) Dequeue() (T, bool) {
	if len(q.data) == 0 {
		var zeroValue T
		return zeroValue, false
	}
	value := q.data[0]
	q.data = q.data[1:]
	return value, true
}

// Peek: 큐의 첫 번째 요소를 반환
func (q *Queue[T]) Peek() (T, bool) {
	if len(q.data) == 0 {
		var zeroValue T
		return zeroValue, false
	}
	return q.data[0], true
}

// IsEmpty: 큐가 비었는지 확인
func (q *Queue[T]) IsEmpty() bool {
	return len(q.data) == 0
}

// Size: 큐의 크기 반환
func (q *Queue[T]) Size() int {
	return len(q.data)
}

func main() {
	// int 타입 큐 생성
	q := NewQueue[int]()

	// Enqueue 테스트
	q.Enqueue(10)
	q.Enqueue(20)
	q.Enqueue(30)

	fmt.Println("큐 상태:", q.data) // 출력: 큐 상태: [10 20 30]

	// Dequeue 테스트
	value, ok := q.Dequeue()
	if ok {
		fmt.Println("Dequeue 값:", value) // 출력: Dequeue 값: 10
	}

	// Peek 테스트
	peekValue, ok := q.Peek()
	if ok {
		fmt.Println("Peek 값:", peekValue) // 출력: Peek 값: 20
	}

	// 큐 상태 확인
	fmt.Println("큐 상태:", q.data) // 출력: 큐 상태: [20 30]

	// 큐 비우기
	q.Dequeue()
	q.Dequeue()
	fmt.Println("큐 비었는지:", q.IsEmpty()) // 출력: 큐 비었는지: true
}
