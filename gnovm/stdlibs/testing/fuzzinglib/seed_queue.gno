package fuzzinglib

import "errors"

// Seed_Queue 구조체 정의
type Seed_Queue struct {
	data []Seed
}

// New_Seed_Queue: 새로운 Seed_Queue 생성
func New_Seed_Queue() *Seed_Queue {
	return &Seed_Queue{
		data: make([]Seed, 0), // 빈 슬라이스로 초기화
	}
}

// Enqueue: 큐에 Seed 추가
func (q *Seed_Queue) Enqueue(seed Seed) {
	q.data = append(q.data, seed)
}

// Dequeue: 큐에서 첫 번째 Seed 제거 및 반환
func (q *Seed_Queue) Dequeue() (Seed, bool) {
	if len(q.data) == 0 {
		return Seed{}, false // 빈 큐인 경우
	}
	// 첫 번째 원소를 가져옴
	front := q.data[0]
	// 슬라이스에서 제거
	q.data = q.data[1:]
	// 메모리 정리 로직: 슬라이스 길이가 용량의 절반 이하일 경우 새로운 슬라이스 생성
	if len(q.data) > 0 && len(q.data) <= cap(q.data)/2 {
		newData := make([]Seed, len(q.data))
		copy(newData, q.data)
		q.data = newData
	}
	return front, true
}

// Peek: 큐의 첫 번째 Seed 조회
func (q *Seed_Queue) Peek() (Seed, error) {
	if len(q.data) == 0 {
		return Seed{}, errors.New("queue is empty")
	}
	return q.data[0], nil
}

// IsEmpty: 큐가 비어있는지 확인
func (q *Seed_Queue) IsEmpty() bool {
	return len(q.data) == 0
}

// Size: 큐의 크기 반환
func (q *Seed_Queue) Size() int {
	return len(q.data)
}

func (q *Seed_Queue) Display() {
	if len(q.data) == 0 {
		println("Queue is empty")
		return
	}
	println("Queue seeds:")
	for i, seed := range q.data {
		println("[", i, "]: {content:", seed.Content, "}")
	}
}
