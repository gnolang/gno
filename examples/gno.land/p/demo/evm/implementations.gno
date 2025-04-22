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
