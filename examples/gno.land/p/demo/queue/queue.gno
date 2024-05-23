package queue

// Queue is a queue data structure that stores elements of any type.
type Queue struct {
	begin uint64                 // Index of the front element
	end   uint64                 // Index after the last element
	data  map[uint64]interface{} // Storage for queue elements
}

// NewQueue creates and initializes a new queue.
func NewQueue() *Queue {
	return &Queue{
		begin: 0,
		end:   0,
		data:  make(map[uint64]interface{}),
	}
}

// PushBack adds an element to the end of the queue.
// Returns an error if the queue is full.
func (q *Queue) PushBack(value interface{}) error {
	if q.end+1 == q.begin {
		return ErrResourceError
	}

	q.data[q.end] = value
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
	value := q.data[q.end]
	delete(q.data, q.end)

	return value, nil
}

// PushFront adds an element to the front of the queue.
// Returns an error if the queue is full.
func (q *Queue) PushFront(value interface{}) error {
	if q.begin-1 == q.end {
		return ErrResourceError
	}

	q.begin--
	q.data[q.begin] = value

	return nil
}

// PopFront removes and returns the element at the front of the queue.
// Returns an error if the queue is empty.
func (q *Queue) PopFront() (interface{}, error) {
	if q.begin == q.end {
		return nil, ErrEmptyArrayPop
	}

	value := q.data[q.begin]
	delete(q.data, q.begin)
	q.begin++

	return value, nil
}

// Front returns the element at the front of the queue without removing it.
// Returns an error if the queue is empty.
func (q *Queue) Front() (interface{}, error) {
	if q.Empty() {
		return nil, ErrArrayOutOfBounds
	}

	return q.data[q.begin], nil
}

// Back returns the element at the end of the queue without removing it.
// Returns an error if the queue is empty.
func (q *Queue) Back() (interface{}, error) {
	if q.Empty() {
		return nil, ErrArrayOutOfBounds
	}

	return q.data[q.end-1], nil
}

// At returns the element at the specified index in the queue.
// Returns an error if the index is out of bounds.
func (q *Queue) At(index uint64) (interface{}, error) {
	if index >= q.Length() {
		return nil, ErrArrayOutOfBounds
	}

	return q.data[q.begin+index], nil
}

// Clear removes all elements from the queue.
func (q *Queue) Clear() {
	q.begin = 0
	q.end = 0
	q.data = make(map[uint64]interface{})
}

// Length returns the number of elements in the queue.
func (q *Queue) Length() uint64 {
	return q.end - q.begin
}

// Empty returns true if the queue is empty, false otherwise.
func (q *Queue) Empty() bool {
	return q.end == q.begin
}
