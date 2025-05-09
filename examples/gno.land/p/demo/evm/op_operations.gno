package evm

import (
	"errors"

	"gno.land/p/demo/uint256"
)

func opOperation(f func(a, b *uint256.Uint) *uint256.Uint) func(s *State) error {
	return func(s *State) error {
		a, err := s.stack.Pop()
		if err != nil {
			return errors.New("can't do operation: empty stack")
		}
		b, err := s.stack.Pop()
		if err != nil {
			return errors.New("can't do operation: stack of size 1")
		}
		s.stack.Push(f(a, b))
		return nil
	}
}

func opExp(s *State) error {
	a, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't do operation: empty stack")
	}
	b, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't do operation: stack of size 1")
	}
	s.stack.Push(a.Exp(a, b))
	s.gasConsumed += 50 * uint((b.BitLen()+7)/8)
	return nil
}

func opAddmod(s *State) error {
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

func opMulmod(s *State) error {
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

func opNot(s *State) error {
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
func opByte(s *State) error {
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

func opSlt(s *State) error {
	y, err := s.stack.Pop()
	if err != nil {
		return err
	}
	x, err := s.stack.Pop()
	if err != nil {
		return err
	}
	xSign := x.Sign()
	ySign := y.Sign()
	if xSign != ySign {
		if xSign < ySign { // if x and y have different signs
			s.stack.Push(uint256.One())
		} else {
			s.stack.Push(uint256.Zero())
		}
		return nil
	}
	if x.Cmp(y) < 0 { // same sign & compare as unsigned
		s.stack.Push(uint256.One())
	} else {
		s.stack.Push(uint256.Zero())
	}
	return nil
}

func opSgt(s *State) error {
	y, err := s.stack.Pop()
	if err != nil {
		return err
	}
	x, err := s.stack.Pop()
	if err != nil {
		return err
	}
	xSign := x.Sign()
	ySign := y.Sign()
	if xSign != ySign {
		if xSign > ySign {
			s.stack.Push(uint256.One())
		} else {
			s.stack.Push(uint256.Zero())
		}
		return nil
	}
	if x.Cmp(y) > 0 {
		s.stack.Push(uint256.One())
	} else {
		s.stack.Push(uint256.Zero())
	}
	return nil
}

func opSar(s *State) error {
	shift, err := s.stack.Pop()
	if err != nil {
		return err
	}
	value, err := s.stack.Pop()
	if err != nil {
		return err
	}
	if shift.GtUint64(256) {
		if value.Sign() >= 0 {
			value.Clear()
		} else {
			// Max negative shift: all bits set
			value.SetAllOne()
		}
		s.stack.Push(value)
		return nil
	}
	n := uint(shift.Uint64())
	s.stack.Push(value.SRsh(value, n))
	return nil
}

func opReturnDataSize(s *State) error {
	s.stack.Push(uint256.NewUint(uint64(len(s.result))))
	return nil
}
