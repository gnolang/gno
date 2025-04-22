package evm

import (
	"errors"
	"strconv"
)

var outOfBounds = errors.New("out of bounds")

func opNotImplemented(s *State) error {
	return errors.New("op code " + strconv.FormatInt(int64(s.code.Current()), 16) + ": not implemented")
}

func opAdd(s *State) error {
	a1, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't ADD: empty stack")
	}

	a2, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't ADD: stack of size 1")
	}

	s.stack.Push(a1 + a2)
	s.code.Increment()
	return nil
}

func opMul(s *State) error {
	a1, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't MUL: empty stack")
	}

	a2, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't MUL: stack of size 1")
	}

	s.stack.Push(a1 * a2)
	s.code.Increment()
	return nil
}

func opPush1(s *State) error {
	s.code.Increment()
	if s.code.Ill() {
		return outOfBounds
	}
	x := s.code.Read(1)
	s.stack.Push(byte(x))
	return nil
}
