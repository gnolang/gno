package stack

type Stack struct {
	top    *node
	length int
}

type node struct {
	value interface{}
	prev  *node
}

func New() *Stack {
	return &Stack{nil, 0}
}

func (s *Stack) Len() int {
	return s.length
}

func (s *Stack) Top() interface{} {
	if s.length == 0 {
		return nil
	}
	return s.top.value
}

func (s *Stack) Pop() interface{} {
	if s.length == 0 {
		return nil
	}

	node := s.top
	s.top = node.prev
	s.length -= 1
	return node.value
}

func (s *Stack) Push(value interface{}) {
	node := &node{value, s.top}
	s.top = node
	s.length += 1
}
