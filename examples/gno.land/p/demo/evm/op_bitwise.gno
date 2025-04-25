package evm

import (
	"errors"

	"gno.land/p/demo/uint256"
)

func opAnd(s *state) error {
    a, err := s.stack.Pop()
    if err != nil {
        return errors.New("can't AND: empty stack")
    }
    b, err := s.stack.Pop()
    if err != nil {
        return errors.New("can't AND: stack of size 1")
    }
    s.stack.Push(a.And(a, b))
    return nil
}

func opOr(s *state) error {
    a, err := s.stack.Pop()
    if err != nil {
        return errors.New("can't OR: empty stack")
    }
    b, err := s.stack.Pop()
    if err != nil {
        return errors.New("can't OR: stack of size 1")
    }
    s.stack.Push(a.Or(a, b))
    return nil
}

func opXor(s *state) error {
    a, err := s.stack.Pop()
    if err != nil {
        return errors.New("can't XOR: empty stack")
    }
    b, err := s.stack.Pop()
    if err != nil {
        return errors.New("can't XOR: stack of size 1")
    }
    s.stack.Push(a.Xor(a, b))
    return nil
}

func opNot(s *state) error {
    a, err := s.stack.Pop()
    if err != nil {
        return errors.New("can't NOT: empty stack")
    }
    s.stack.Push(a.Not(a))
    return nil
}

// Retrieve a byte from a word.
// IN: `[offset, number]`
// OUT: `[byte]`
func opByte(s *state) error {
	off, err := s.stack.Pop()
    if err != nil {
        return errors.New("can't XOR: empty stack")
    }
    number, err := s.stack.Pop()
    if err != nil {
        return errors.New("can't XOR: stack of size 1")
    }
	off64 := off.Uint64()
	bytes := number.Bytes32()

	var res *uint256.Uint
	if off64 >= 32 {
		res = uint256.Zero()
	} else {
		res = uint256.NewUint(uint64(bytes[off64]))
	}
	s.stack.Push(res)
	return nil
}