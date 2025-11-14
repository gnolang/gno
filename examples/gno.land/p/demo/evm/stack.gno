package evm

import (
	"errors"

	"gno.land/p/onbloc/uint256"
)

type evmStack struct {
	data   []*uint256.Uint
	length uint
}

func newEvmStack() *evmStack {
	return &evmStack{
		data:   make([]*uint256.Uint, 0),
		length: 0,
	}
}

func (s evmStack) print() (res string) {
	if s.length == 0 {
		res = "stack is empty!\n"
	}
	for _, e := range s.data {
		res += e.String() + "\n"
	}
	return
}

func (s *evmStack) push(b *uint256.Uint) {
	s.data = append(s.data, b)
	s.length++
}

func (s *evmStack) pop() (*uint256.Uint, error) {
	if s.length == 0 {
		return uint256.Zero(), errors.New("stack is empty")
	}
	top := s.data[s.length-1]
	s.data = s.data[:s.length-1]
	s.length--
	return top, nil
}

func (s *evmStack) top() (*uint256.Uint, error) {
	if s.length == 0 {
		return nil, errors.New("stack is empty")
	}
	return s.data[s.length-1], nil
}

func (s *evmStack) peek(n int) (*uint256.Uint, error) {
	if s.length == 0 {
		return nil, errors.New("stack is empty")
	}
	if n >= int(s.length) {
		return nil, errors.New("stack access out of bounds")
	}
	return s.data[int(s.length)-1-n], nil
}

func (s *evmStack) len() int {
	return len(s.data)
}

func (s evmStack) String() string {
	if s.length == 0 {
		return "[]"
	}
	start := int(s.length - 1)
	buf := "["
	for i := start; i >= 0; i-- {
		if i != start {
			buf += ", "
		}
		buf += s.data[i].Hex()
	}
	return buf + "]"
}
