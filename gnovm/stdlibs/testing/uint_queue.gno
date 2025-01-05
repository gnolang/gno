package testing

type Uint_Queue struct {
	data []uint
}

func New_Uint_Queue() *Uint_Queue {
	return &Uint_Queue{data: make([]uint, 0)}
}

func (q *Uint_Queue) Uint_Enqueue(value uint) {
	q.data = append(q.data, value)
}

// Remove from queue and return.
// If the slice is empty compared to the capacity,
// Added logic to copy into new slices to organize memory.
func (q *Uint_Queue) Uint_Dequeue() (uint, bool) {
	if len(q.data) == 0 {
		return 0, false
	}

	value := q.data[0]
	q.data[0] = 0

	q.data = q.data[1:]

	// Memory leak prevention
	// When the current slice length (len) is less than half the capacity (cap)
	// Copy to new slice, remove the front discarded space.
	if len(q.data) > 0 && len(q.data) <= cap(q.data)/2 {
		newData := make([]uint, len(q.data))
		copy(newData, q.data)
		q.data = newData
	}

	return value, true
}

func (q *Uint_Queue) Uint_Peek() (uint, bool) {
	if len(q.data) == 0 {
		return 0, false
	}
	return q.data[0], true
}

func (q *Uint_Queue) IsEmpty() bool {
	return len(q.data) == 0
}

func (q *Uint_Queue) Size() int {
	return len(q.data)
}
