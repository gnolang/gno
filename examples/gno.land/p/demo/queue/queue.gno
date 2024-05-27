package queue

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
)

// Queue is a queue data structure that stores elements of any type.
type Queue struct {
	begin seqid.ID  // Index of the front element
	end   seqid.ID  // Index after the last element
	data  *avl.Tree // Storage for queue elements
}

// NewQueue creates and initializes a new queue.
func NewQueue() *Queue {
	return &Queue{
		begin: seqid.ID(0),
		end:   seqid.ID(0),
		data:  avl.NewTree(),
	}
}

// PushBack adds an element to the end of the queue.
// Returns an error if the queue is full.
func (q *Queue) PushBack(value interface{}) error {
	if q.end+1 == q.begin {
		return ErrResourceError
	}

	q.data.Set(q.end.String(), value)

	q.end++

	return nil
}

// PopBack removes and returns the element at the end of the queue.
// Returns an error if the queue is empty.
func (q *Queue) PopBack() (interface{}, error) {
	if q.begin == q.end {
		return nil, ErrEmptyArrayPop
	}

	q.end--
	value, ok := q.data.Get(q.end.String())
	if !ok {
		return nil, ErrNonExistedValue
	}

	q.data.Remove(q.end.String())

	return value, nil
}

// PushFront adds an element to the front of the queue.
// Returns an error if the queue is full.
func (q *Queue) PushFront(value interface{}) error {
	if q.begin-1 == q.end {
		return ErrResourceError
	}

	q.begin--
	q.data.Set(q.begin.String(), value)

	return nil
}

// PopFront removes and returns the element at the front of the queue.
// Returns an error if the queue is empty.
func (q *Queue) PopFront() (interface{}, error) {
	if q.begin == q.end {
		return nil, ErrEmptyArrayPop
	}

	value, ok := q.data.Get(q.begin.String())
	if !ok {
		return nil, ErrNonExistedValue
	}

	q.data.Remove(q.begin.String())

	q.begin++

	return value, nil
}

// Front returns the element at the front of the queue without removing it.
// Returns an error if the queue is empty.
func (q *Queue) Front() (interface{}, error) {
	if q.Empty() {
		return nil, ErrArrayOutOfBounds
	}

	id := q.begin

	value, ok := q.data.Get(id.String())
	if !ok {
		return nil, ErrNonExistedValue
	}

	return value, nil
}

// Back returns the element at the end of the queue without removing it.
// Returns an error if the queue is empty.
func (q *Queue) Back() (interface{}, error) {
	if q.Empty() {
		return nil, ErrArrayOutOfBounds
	}

	id := q.end - 1

	value, ok := q.data.Get(id.String())
	if !ok {
		return nil, ErrNonExistedValue
	}

	return value, nil
}

// At returns the element at the specified index in the queue.
// Returns an error if the index is out of bounds.
func (q *Queue) At(index uint64) (interface{}, error) {
	if index >= q.Length() {
		return nil, ErrArrayOutOfBounds
	}

	id := q.begin + seqid.ID(index)

	value, ok := q.data.Get(id.String())
	if !ok {
		return nil, ErrNonExistedValue
	}

	return value, nil
}

// Clear removes all elements from the queue.
func (q *Queue) Clear() {
	q.begin = 0
	q.end = 0
	q.data = avl.NewTree()
}

// Length returns the number of elements in the queue.
func (q *Queue) Length() uint64 {
	return uint64(q.end - q.begin)
}

// Empty returns true if the queue is empty, false otherwise.
func (q *Queue) Empty() bool {
	return q.end == q.begin
}
