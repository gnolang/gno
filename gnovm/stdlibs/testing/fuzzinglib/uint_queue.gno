package fuzzinglib

// Queue 구조체
type Uint_Queue struct {
	data []uint
}

// NewQueue: 새로운 큐 생성
func New_Uint_Queue() *Uint_Queue {
	return &Uint_Queue{data: make([]uint, 0)}
}

// EnUint_Queue: 큐에 삽입
func (q *Uint_Queue) Uint_Enqueue(value uint) {
	q.data = append(q.data, value)
}

// Dequeue: 큐에서 제거 후 반환.
func (q *Uint_Queue) Uint_Dequeue() (uint, bool) {
	if len(q.data) == 0 {
		return 0, false
	}
	value := q.data[0]
	q.data = q.data[1:]
	return value, true
}

// Peek: 큐의 첫 번째 요소를 반환
func (q *Uint_Queue) Uint_Peek() (uint, bool) {
	if len(q.data) == 0 {
		return 0, false
	}
	return q.data[0], true
}

// IsEmpty: 큐가 비었는지 확인
func (q *Uint_Queue) IsEmpty() bool {
	return len(q.data) == 0
}

// Size: 큐의 크기 반환
func (q *Uint_Queue) Size() int {
	return len(q.data)
}

// func main() {
// 	// Queue 생성
// 	q := New_Uint_Queue()

// 	// 테스트
// 	q.Uint_Enqueue(10)
// 	q.Uint_Enqueue(20)
// 	q.Uint_Enqueue(30)
// 	println("큐 상태:", q.data)

// 	value, ok := q.Uint_Dequeue()
// 	if ok {
// 		println("Dequeue 값:", value)
// 	}

// 	peekValue, ok := q.Uint_Peek()
// 	if ok {
// 		println("Peek 값:", peekValue)
// 	}

// 	println("큐 상태:", q.data)
// }
