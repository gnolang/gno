package evm

import "errors"

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
    s.code.Increment()
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
    s.code.Increment()
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
    s.code.Increment()
    return nil
}

func opNot(s *state) error {
    a, err := s.stack.Pop()
    if err != nil {
        return errors.New("can't NOT: empty stack")
    }
    s.stack.Push(a.Not(a))
    s.code.Increment()
    return nil
}