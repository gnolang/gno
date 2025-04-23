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
			return errors.New("can't PUSH" + strconv.Itoa(n) + ": " + outOfBounds.Error())
		}
		stackData, err := s.code.Read(n)
		if err != nil {
			return errors.New("can't PUSH" + strconv.Itoa(n) + ": " + err.Error())
		}
		s.stack.Push(stackData)
		return nil
	}
}

func opPop(s *state) error {
	_, err := s.stack.Pop()
	return err
}

func opJump(s *state) error {
	dest, err := s.stack.Pop()
	if err != nil {
		return err
	}
	inc := int(dest.To64())
	opDest := OpCode(s.code.At(inc))
	if opDest != JUMPDEST {
		return errors.New("can't JUMP: JUMP needs to land on a JUMPDEST, but here, it's a 0x" + strconv.FormatUint(uint64(opDest), 16) + ".")
	}
	s.code.rip = inc + 1
	return nil
}

func opNop(s *state) error {
	s.code.Increment()
	return nil
}

func opJumpi(s *state) error {
	dest, err := s.stack.Pop()
	if err != nil {
		return err
	}
	doJump, err := s.stack.Pop()
	if err != nil {
		return err
	}
	if doJump.IsZero() {
		s.code.Increment()
		return nil
	}
	inc := int(dest.To64())
	opDest := OpCode(s.code.At(inc))
	if opDest != JUMPDEST {
		return errors.New("can't JUMP: JUMP needs to land on a JUMPDEST, but here, it's a 0x" + strconv.FormatUint(uint64(opDest), 16) + ".")
	}
	s.code.rip = inc + 1
	return nil
}

func opIszero(s *state) error {
	num, err := s.stack.Pop()
	if err != nil {
		return err
	}
	if num.IsZero() {
		s.stack.Push(uint256.NewFrom64s(1))
	} else {
		s.stack.Push(uint256.New())
	}
	s.code.Increment()
	return nil
}