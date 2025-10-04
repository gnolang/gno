package evm

import (
	"errors"

	"gno.land/p/onbloc/uint256"
)

func opOperation(f func(a, b *uint256.Uint) *uint256.Uint) func(s *state) error {
	return func(s *state) error {
		a, err := s.stack.pop()
		if err != nil {
			return errors.New("can't do operation: empty stack")
		}
		b, err := s.stack.pop()
		if err != nil {
			return errors.New("can't do operation: stack of size 1")
		}
		s.stack.push(f(a, b).Clone())
		return nil
	}
}

func opExp(s *state) error {
	a, err := s.stack.pop()
	if err != nil {
		return errors.New("can't do operation: empty stack")
	}
	b, err := s.stack.pop()
	if err != nil {
		return errors.New("can't do operation: stack of size 1")
	}
	s.stack.push(a.Exp(a, b).Clone())
	s.gasConsumed += 50 * uint((b.BitLen()+7)/8)
	return nil
}

func opAddmod(s *state) error {
	a, err := s.stack.pop()
	if err != nil {
		return errors.New("can't ADDMOD: empty stack")
	}

	b, err := s.stack.pop()
	if err != nil {
		return errors.New("can't ADDMOD: stack of size 1")
	}

	n, err := s.stack.pop()
	if err != nil {
		return errors.New("can't ADDMOD: stack of size 2")
	}

	s.stack.push(a.Add(a, b).Mod(a, n))
	return nil
}

func opMulmod(s *state) error {
	a, err := s.stack.pop()
	if err != nil {
		return errors.New("can't MULMOD: empty stack")
	}

	b, err := s.stack.pop()
	if err != nil {
		return errors.New("can't MULMOD: stack of size 1")
	}

	n, err := s.stack.pop()
	if err != nil {
		return errors.New("can't MULMOD: stack of size 2")
	}

	s.stack.push(a.MulMod(a, b, n))
	return nil
}

func opNot(s *state) error {
	a, err := s.stack.pop()
	if err != nil {
		return errors.New("can't NOT: empty stack")
	}
	s.stack.push(a.Not(a))
	return nil
}

// Retrieve a byte from a word.
// IN: `[offset, number]`
// OUT: `[byte]`
func opByte(s *state) error {
	off, err := s.stack.pop()
	if err != nil {
		return errors.New("can't XOR: empty stack")
	}
	number, err := s.stack.pop()
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
	s.stack.push(res)
	return nil
}

func opSlt(s *state) error {
	y, err := s.stack.pop()
	if err != nil {
		return err
	}
	x, err := s.stack.pop()
	if err != nil {
		return err
	}
	xSign := x.Sign()
	ySign := y.Sign()
	if xSign != ySign {
		if xSign < ySign { // if x and y have different signs
			s.stack.push(uint256.One())
		} else {
			s.stack.push(uint256.Zero())
		}
		return nil
	}
	if x.Cmp(y) < 0 { // same sign & compare as unsigned
		s.stack.push(uint256.One())
	} else {
		s.stack.push(uint256.Zero())
	}
	return nil
}

func opSgt(s *state) error {
	y, err := s.stack.pop()
	if err != nil {
		return err
	}
	x, err := s.stack.pop()
	if err != nil {
		return err
	}
	xSign := x.Sign()
	ySign := y.Sign()
	if xSign != ySign {
		if xSign > ySign {
			s.stack.push(uint256.One())
		} else {
			s.stack.push(uint256.Zero())
		}
		return nil
	}
	if x.Cmp(y) > 0 {
		s.stack.push(uint256.One())
	} else {
		s.stack.push(uint256.Zero())
	}
	return nil
}

func opSar(s *state) error {
	shift, err := s.stack.pop()
	if err != nil {
		return err
	}
	value, err := s.stack.pop()
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
		s.stack.push(value.Clone())
		return nil
	}
	n := uint(shift.Uint64())
	s.stack.push(value.SRsh(value, n))
	return nil
}
