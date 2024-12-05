package main

// Queue 구조체
type Int_Queue struct {
	data []int
}

// NewQueue: 새로운 큐 생성
func New_Int_Queue() *Int_Queue {
	return &Int_Queue{data: make([]int, 0)}
}

// EnInt_Queue: 큐에 삽입
func (q *Int_Queue) Int_Enqueue(value int) {
	q.data = append(q.data, value)
}

// Dequeue: 큐에서 제거 후 반환
func (q *Int_Queue) Int_Dequeue() (int, bool) {
	if len(q.data) == 0 {
		return 0, false
	}
	value := q.data[0]
	q.data = q.data[1:]
	return value, true
}

// Peek: 큐의 첫 번째 요소를 반환
func (q *Int_Queue) Int_Peek() (int, bool) {
	if len(q.data) == 0 {
		return 0, false
	}
	return q.data[0], true
}

// IsEmpty: 큐가 비었는지 확인
func (q *Int_Queue) IsEmpty() bool {
	return len(q.data) == 0
}

// Size: 큐의 크기 반환
func (q *Int_Queue) Size() int {
	return len(q.data)
}

// func main() {
// 	// Queue 생성
// 	q := New_Int_Queue()

// 	// 테스트
// 	q.Int_Enqueue(10)
// 	q.Int_Enqueue(20)
// 	q.Int_Enqueue(30)

// 	println("큐 상태:", q.data)

// 	value, ok := q.Int_Dequeue()
// 	if ok {
// 		println("Dequeue 값:", value)
// 	}

// 	peekValue, ok := q.Int_Peek()
// 	if ok {
// 		println("Peek 값:", peekValue)
// 	}

// 	println("큐 상태:", q.data)
// }
