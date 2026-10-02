package mathml

type stack[T any] struct {
	data []T
	top  int
}

// Create a new stack
func newStack[T any]() *stack[T] {
	return &stack[T]{
		data: make([]T, 0),
		top:  -1,
	}
}

// Push a value onto the stack
func (s *stack[T]) Push(val T) {
	s.top++
	if len(s.data) <= s.top { // Check if we need to grow the slice
		newSize := len(s.data) * 2
		if newSize == 0 {
			newSize = 1 // Start with a minimum capacity if the stack is empty
		}
		newData := make([]T, newSize)
		copy(newData, s.data) // Copy old elements to new slice
		s.data = newData
	}
	s.data[s.top] = val
}

// Peek the top element from the stack
func (s *stack[T]) Peek() (val T) {
	val = s.data[s.top]
	return
}

// Pop the top element from the stack
func (s *stack[T]) Pop() (val T) {
	val = s.data[s.top]
	s.top--
	return
}

// Check if the stack is empty
func (s *stack[T]) empty() bool {
	return s.top < 0
}
