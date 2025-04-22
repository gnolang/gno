package evm

import (
	"errors"
	"strconv"

	"gno.land/p/demo/evm/uint256"
)

var outOfBounds = errors.New("out of bounds")

func opNotImplemented(s *state) error {
	return errors.New("op code " + strconv.FormatInt(int64(s.code.Current()), 16) + ": not implemented")
}

func opAdd(s *state) error {
	a1, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't ADD: empty stack")
	}

	a2, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't ADD: stack of size 1")
	}

	s.stack.Push(a1.Add256(a2))
	s.code.Increment()
	return nil
}

func opMul(s *state) error {
	a1, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't MUL: empty stack")
	}

	a2, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't MUL: stack of size 1")
	}

	s.stack.Push(a1.Mul256(a2))
	s.code.Increment()
	return nil
}

func opPush0(s *state) error {
	s.stack.Push(uint256.New())
	s.code.Increment()
	return nil
}

func opPush(n int) func(*state) error {
	return func(s *state) error {
		s.code.Increment()
		if s.code.Ill() {
			return outOfBounds
		}
		s.stack.Push(s.code.Read(n))
		return nil
	}
}
