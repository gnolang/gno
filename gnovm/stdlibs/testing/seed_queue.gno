package testing

import "errors"

type Seed_Queue struct {
	data []Seed
}

func New_Seed_Queue() *Seed_Queue {
	return &Seed_Queue{
		data: make([]Seed, 0),
	}
}

func (q *Seed_Queue) Enqueue(seed Seed) {
	q.data = append(q.data, seed)
}

func (q *Seed_Queue) Dequeue() (Seed, bool) {
	if len(q.data) == 0 {
		return Seed{}, false
	}

	front := q.data[0]

	q.data = q.data[1:]
	// Reduce memory usage
	if len(q.data) > 0 && len(q.data) <= cap(q.data)/2 {
		newData := make([]Seed, len(q.data))
		copy(newData, q.data)
		q.data = newData
	}
	return front, true
}

func (q *Seed_Queue) Peek() (Seed, error) {
	if len(q.data) == 0 {
		return Seed{}, errors.New("queue is empty")
	}
	return q.data[0], nil
}

func (q *Seed_Queue) Select() Seed

func (q *Seed_Queue) IsEmpty() bool {
	return len(q.data) == 0
}

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
