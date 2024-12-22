package fuzzinglib

type (
	gen uint
	// Trio는 (uint, uint, uint) 삼중쌍을 나타냅니다.
	Trio struct {
		First          uint
		Second         uint
		NextStartPoint uint
	}
)

// Uint_Trio_Queue 구조체
type Uint_Trio_Queue struct {
	data []Trio
}

// New_Uint_Trio_Queue: 새로운 큐 생성
func New_Uint_Trio_Queue() *Uint_Trio_Queue {
	return &Uint_Trio_Queue{data: make([]Trio, 0)}
}

// Uint_Trio_Enqueue: 큐에 Trio 삽입
func (q *Uint_Trio_Queue) Uint_Trio_Enqueue(trio Trio) {
	q.data = append(q.data, trio)
}

// Uint_Trio_Dequeue: 큐에서 Trio 제거 후 반환
// 슬라이스가 용량(cap)에 비해 많이 비어 있을 경우,
// 새로운 슬라이스로 복사해 메모리를 정리하는 로직을 추가하였습니다.
func (q *Uint_Trio_Queue) Uint_Trio_Dequeue() (Trio, bool) {
	if len(q.data) == 0 {
		return Trio{}, false
	}

	value := q.data[0]
	// 사용된 첫 요소를 GC가 수거할 수 있도록 0값으로 세팅(선택 사항)
	q.data[0] = Trio{}

	// 맨 앞 요소를 제외
	q.data = q.data[1:]

	// 메모리 누수 방지 로직:
	// 슬라이스 길이가 용량의 절반 이하라면 앞부분을 버리고 새 슬라이스로 복사
	if len(q.data) > 0 && len(q.data) <= cap(q.data)/2 {
		newData := make([]Trio, len(q.data))
		copy(newData, q.data)
		q.data = newData
	}

	return value, true
}

// Uint_Trio_Peek: 큐의 첫 번째 요소 반환
func (q *Uint_Trio_Queue) Uint_Trio_Peek() (Trio, bool) {
	if len(q.data) == 0 {
		return Trio{}, false
	}
	return q.data[0], true
}

// IsEmpty: 큐가 비었는지 확인
func (q *Uint_Trio_Queue) IsEmpty() bool {
	return len(q.data) == 0
}

// Size: 큐의 크기 반환
func (q *Uint_Trio_Queue) Size() int {
	return len(q.data)
}

// Uint_Trio_Queue_List 구조체 (예시로 존재한다고 가정)
type Uint_Trio_Queue_List struct {
	queues []*Uint_Trio_Queue
}

// New_Uint_Trio_Queue_List: 새로운 큐 리스트 생성 (예시)
func New_Uint_Trio_Queue_List() *Uint_Trio_Queue_List {
	return &Uint_Trio_Queue_List{queues: make([]*Uint_Trio_Queue, 0)}
}

// AddQueue: 큐 리스트에 큐를 추가 (예시)
func (ql *Uint_Trio_Queue_List) AddQueue(q *Uint_Trio_Queue) {
	ql.queues = append(ql.queues, q)
}

// 예시 main 함수
// func main() {
// 	// 새 큐 생성
// 	queue := New_Uint_Trio_Queue()
//
// 	// 큐에 Trio 삽입
// 	queue.Uint_Trio_Enqueue(Trio{First: 1, Second: 2, NextStartPoint: 3})
// 	queue.Uint_Trio_Enqueue(Trio{First: 4, Second: 5, NextStartPoint: 6})
// 	queue.Uint_Trio_Enqueue(Trio{First: 7, Second: 8, NextStartPoint: 9})
//
// 	// 큐 출력
// 	println("Queue size:", queue.Size())
// 	for !queue.IsEmpty() {
// 		trio, _ := queue.Uint_Trio_Dequeue()
// 		println("Dequeued: {First: %d, Second: %d, NextStartPoint: %d}\n",
// 			trio.First, trio.Second, trio.NextStartPoint)
// 	}
//
// 	// 큐 리스트 테스트
// 	queueList := New_Uint_Trio_Queue_List()
// 	queueList.AddQueue(queue)
// 	println("Queue list size:", len(queueList.queues))
// }
