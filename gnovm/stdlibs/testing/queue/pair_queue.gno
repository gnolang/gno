package main

type (
	gen uint
	// Pair는 (uint, uint) 쌍을 나타냅니다.
	Pair struct {
		First  uint
		Second uint
	}
)

// Uint_Pair_Queue 구조체
type Uint_Pair_Queue struct {
	data []Pair
}

// New_Unit_Pair_Queue: 새로운 큐 생성
func New_Uint_Pair_Queue() *Uint_Pair_Queue {
	return &Uint_Pair_Queue{data: make([]Pair, 0)}
}

// Unit_Pair_Enqueue: 큐에 삽입
func (q *Uint_Pair_Queue) Uint_Pair_Enqueue(pair Pair) {
	q.data = append(q.data, pair)
}

// Unit_Pair_Dequeue: 큐에서 제거 후 반환
func (q *Uint_Pair_Queue) Uint_Pair_Dequeue() (Pair, bool) {
	if len(q.data) == 0 {
		return Pair{}, false
	}
	value := q.data[0]
	q.data = q.data[1:]
	return value, true
}

// Unit_Pair_Peek: 큐의 첫 번째 요소를 반환
func (q *Uint_Pair_Queue) Uint_Pair_Peek() (Pair, bool) {
	if len(q.data) == 0 {
		return Pair{}, false
	}
	return q.data[0], true
}

// IsEmpty: 큐가 비었는지 확인
func (q *Uint_Pair_Queue) IsEmpty() bool {
	return len(q.data) == 0
}

// Size: 큐의 크기 반환
func (q *Uint_Pair_Queue) Size() int {
	return len(q.data)
}

// Uint_Pair_Queue_List 구조체: 큐의 리스트
type Uint_Pair_Queue_List struct {
	queues []*Uint_Pair_Queue
}

// New_Uint_Pair_Queue_List: 새로운 큐 리스트 생성
func New_Uint_Pair_Queue_List() *Uint_Pair_Queue_List {
	return &Uint_Pair_Queue_List{queues: make([]*Uint_Pair_Queue, 0)}
}

// AddQueue: 큐 리스트에 새로운 큐 추가
func (ql *Uint_Pair_Queue_List) AddQueue(queue *Uint_Pair_Queue) {
	ql.queues = append(ql.queues, queue)
}

// GetQueue: 특정 인덱스의 큐 반환
func (ql *Uint_Pair_Queue_List) GetQueue(index int) (*Uint_Pair_Queue, bool) {
	if index < 0 || index >= len(ql.queues) {
		return nil, false
	}
	return ql.queues[index], true
}

// 메인 함수
func main() {
	// Uint_Pair_Queue_List 생성
	queueList := New_Uint_Pair_Queue_List()

	// 두 개의 큐 생성
	q1 := New_Uint_Pair_Queue()
	q2 := New_Uint_Pair_Queue()

	// 각 큐에 Pair 삽입
	q1.Uint_Pair_Enqueue(Pair{First: 1, Second: 2})
	q1.Uint_Pair_Enqueue(Pair{First: 3, Second: 4})
	q2.Uint_Pair_Enqueue(Pair{First: 5, Second: 6})

	// 큐 리스트에 큐 추가
	queueList.AddQueue(q1)
	queueList.AddQueue(q2)

	// 리스트에서 첫 번째 큐 가져오기
	retrievedQueue, ok := queueList.GetQueue(0)
	if ok {
		println("첫 번째 큐:", retrievedQueue.data)
	}

	// 리스트에서 두 번째 큐 가져오기
	retrievedQueue, ok = queueList.GetQueue(1)
	if ok {
		println("두 번째 큐:", retrievedQueue.data)
	}
}
