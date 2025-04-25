package evm

import (
	"errors"
	"gno.land/p/demo/uint256"
)

func opOperation(f func (a, b *uint256.Uint) *uint256.Uint) func (s *state) error {
	return func(s *state) error {
		a, err := s.stack.Pop()
		if err != nil {
			return errors.New("can't ADD: empty stack")
		}
		b, err := s.stack.Pop()
		if err != nil {
			return errors.New("can't ADD: stack of size 1")
		}
		s.stack.Push(f(a, b))
		return nil
	}
}

func opAddmod(s *state) error {
	a, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't ADDMOD: empty stack")
	}

	b, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't ADDMOD: stack of size 1")
	}

	n, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't ADDMOD: stack of size 2")
	}

	s.stack.Push(a.Add(a, b).Mod(a, n))
	return nil
}

func opMulmod(s *state) error {
	a, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't MULMOD: empty stack")
	}

	b, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't MULMOD: stack of size 1")
	}

	n, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't MULMOD: stack of size 2")
	}

	s.stack.Push(a.MulMod(a, b, n))
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
