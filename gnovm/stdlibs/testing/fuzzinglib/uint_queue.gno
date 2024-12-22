package fuzzinglib

// Queue 구조체
type Uint_Queue struct {
	data []uint
}

// New_Uint_Queue: 새로운 큐 생성
func New_Uint_Queue() *Uint_Queue {
	return &Uint_Queue{data: make([]uint, 0)}
}

// Uint_Enqueue: 큐에 삽입
func (q *Uint_Queue) Uint_Enqueue(value uint) {
	q.data = append(q.data, value)
}

// Uint_Dequeue: 큐에서 제거 후 반환.
// 슬라이스가 용량(cap)에 비해 많이 비어 있을 경우,
// 새로운 슬라이스로 복사해 메모리를 정리하는 로직을 추가함함.
func (q *Uint_Queue) Uint_Dequeue() (uint, bool) {
	if len(q.data) == 0 {
		return 0, false
	}

	value := q.data[0]
	// AI의 말.
	// 필요하다면 q.data[0]을 0으로 세팅해주어 GC가
	// 이 영역을 사용하지 않도록 하는 것도 방법입니다.
	q.data[0] = 0

	q.data = q.data[1:]

	// "메모리 누수" 방지 로직:
	// 현재 슬라이스 길이(len)가 용량(cap)의 절반 이하 시
	// 새 슬라이스에 복사, 앞쪽 버려진 공간을 제거.
	if len(q.data) > 0 && len(q.data) <= cap(q.data)/2 {
		newData := make([]uint, len(q.data))
		copy(newData, q.data)
		q.data = newData
	}

	return value, true
}

// Uint_Peek: 큐의 첫 번째 요소를 반환
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

// 예시 main 함수
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
