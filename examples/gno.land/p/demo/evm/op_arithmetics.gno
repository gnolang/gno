package evm

import (
	"errors"
)

func opAdd(s *state) error {
	a, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't ADD: empty stack")
	}

	b, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't ADD: stack of size 1")
	}

	s.stack.Push(a.Add(a, b))
	return nil
}

func opMul(s *state) error {
	a, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't MUL: empty stack")
	}

	b, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't MUL: stack of size 1")
	}

	s.stack.Push(a.Mul(a, b))
	return nil
}

func opSub(s *state) error {
	a, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't SUB: empty stack")
	}

	b, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't SUB: stack of size 1")
	}

	s.stack.Push(a.Sub(a, b))
	return nil
}

func opDiv(s *state) error {
	a, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't DIV: empty stack")
	}

	b, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't DIV: stack of size 1")
	}
	if b.IsZero() {
		s.stack.Push(b)
	} else {
		s.stack.Push(a.Div(a, b))
	}
	return nil
}

func opSdiv(s *state) error {
	a, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't SDIV: empty stack")
	}

	b, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't SDIV: stack of size 1")
	}

	s.stack.Push(a.SDiv(a, b))
	return nil
}

func opMod(s *state) error {
	a, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't MOD: empty stack")
	}

	b, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't MOD: stack of size 1")
	}
	s.stack.Push(a.Mod(a, b))
	return nil
}

func opSmod(s *state) error {
	a, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't SMOD: empty stack")
	}

	b, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't SMOD: stack of size 1")
	}

	s.stack.Push(a.SMod(a, b))
	return nil
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

func opExp(s *state) error {
	a, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't EXP: empty stack")
	}

	b, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't EXP: stack of size 1")
	}

	s.stack.Push(a.Exp(a, b))
	return nil
}
