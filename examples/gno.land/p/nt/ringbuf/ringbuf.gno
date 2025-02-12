package ringbuf

import (
	"errors"

	"gno.land/p/demo/ufmt"
)

// RingBuffer represents a circular buffer that can store elements of any type.
// It is a fixed-size buffer that overwrites the oldest elements when full.
// It is implemented with interface{} since Gno doesn't support generics yet.
type RingBuffer struct {
	data       []interface{} // Underlying slice to store elements
	size       int           // Fixed size of the buffer
	start, end int           // Indices for the start and end of the buffer
	isFull     bool          // Flag to indicate if the buffer is full
}

var (
	ErrFull        = errors.New("ring buffer is full")
	ErrEmpty       = errors.New("buffer is empty")
	ErrInvalidSize = errors.New("size must be greater than 0")
)

// New creates and initializes a new RingBuffer with the specified size.
// The size must be greater than 0, otherwise an error is returned.
func New(size int) (*RingBuffer, error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}
	return &RingBuffer{
		data: make([]interface{}, size),
		size: size,
	}, nil
}

// Write adds an element to the RingBuffer.
// If the buffer is full, it returns an error.
func (cb *RingBuffer) Write(value interface{}) error {
	if cb.isFull {
		return ErrFull
	}

	cb.data[cb.end] = value
	cb.end = (cb.end + 1) % cb.size

	if cb.end == cb.start {
		cb.isFull = true
	}

	return nil
}

// WriteOverwrite adds an element to the RingBuffer.
// If the buffer is full, it overwrites the oldest element.
func (cb *RingBuffer) WriteOverwrite(value interface{}) {
	if cb.isFull {
		// Buffer is full, overwrite the oldest element
		cb.start = (cb.start + 1) % cb.size
	}

	cb.data[cb.end] = value
	cb.end = (cb.end + 1) % cb.size

	if cb.end == cb.start {
		cb.isFull = true
	}
}

// Read removes and returns the oldest element from the RingBuffer.
// If the buffer is empty, it returns an error.
func (cb *RingBuffer) Read() (interface{}, error) {
	if !cb.isFull && cb.start == cb.end {
		return nil, ErrEmpty
	}

	value := cb.data[cb.start]
	cb.start = (cb.start + 1) % cb.size
	cb.isFull = false

	return value, nil
}

// PrintBufferData prints the current state of the RingBuffer for debugging purposes.
// It displays the buffer's data, start and end indices, and whether it is full.
func (cb *RingBuffer) PrintBufferData() {
	ufmt.Sprintf("Buffer: %v, Start: %d, End: %d, Full: %v\n", cb.data, cb.start, cb.end, cb.isFull)
}
