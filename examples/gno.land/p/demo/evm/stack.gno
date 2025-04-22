package evm

import (
	"errors"

	"gno.land/p/demo/evm/uint256"
)

type evmStack struct {
	data   []*uint256.Uint
	length uint
}

func NewevmStack() *evmStack {
	return &evmStack{
		data:   make([]*uint256.Uint, 0),
		length: 0,
	}
}

func (s evmStack) Print() (res string) {
	if s.length == 0 {
		res = "empty stack!\n"
	}
	for _, e := range s.data {
		res += e.HexString() + "\n"
	}
	return
}

func (s *evmStack) Push(b *uint256.Uint) {
	s.data = append(s.data, b)
	s.length++
}

func (s *evmStack) Pop() (*uint256.Uint, error) {
	if s.length == 0 {
		return uint256.New(), errors.New("stack is empty")
	}
	top := s.data[s.length-1]
	s.data = s.data[:s.length-1]
	s.length--
	return top, nil
}

func (s *evmStack) Top() (*uint256.Uint, error) {
	if s.length == 0 {
		return nil, errors.New("stack is empty")
	}
	return s.data[s.length-1], nil
}
