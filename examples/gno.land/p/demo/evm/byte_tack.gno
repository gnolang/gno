package evm

import "errors"

type ByteStack struct {
	data []byte
}

func NewByteStack() *ByteStack {
	return &ByteStack{
		data: make([]byte, 0),
	}
}

func (s *ByteStack) Push(b byte) {
	s.data = append(s.data, b)
}

func (s *ByteStack) Pop() (byte, error) {
	if len(s.data) == 0 {
		return 0, errors.New("stack underflow")
	}
	top := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return top, nil
}

func (s *ByteStack) Top() (byte, error) {
	if len(s.data) == 0 {
		return 0, errors.New("stack is empty")
	}
	return s.data[len(s.data)-1], nil
}
