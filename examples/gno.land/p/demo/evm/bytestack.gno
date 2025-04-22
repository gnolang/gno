package evm

import "errors"

type byteStack struct {
	data   []byte
	length uint
}

func NewByteStack() *byteStack {
	return &byteStack{
		data:   make([]byte, 0),
		length: 0,
	}
}

func (s *byteStack) Push(b byte) {
	s.data = append(s.data, b)
	s.length++
}

func (s *byteStack) Pop() (byte, error) {
	if s.length == 0 {
		return 0, errors.New("stack is empty")
	}
	top := s.data[s.length-1]
	s.data = s.data[:s.length-1]
	s.length--
	return top, nil
}

func (s *byteStack) Top() (*byte, error) {
	if s.length == 0 {
		return nil, errors.New("stack is empty")
	}
	return &s.data[s.length-1], nil
}
