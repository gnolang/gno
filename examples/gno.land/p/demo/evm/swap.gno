package evm

import (
	"errors"
)

func opSwap1(s *state) error {
	a1, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't SWAP1: empty stack")
	}

	a2, err := s.stack.Pop()
	if err != nil {
		s.stack.Push(a1)
		return errors.New("can't SWAP1: stack of size 1")
	}

	s.stack.Push(a1)
	s.stack.Push(a2)
	s.code.Increment()
	return nil
}

// func opSwap2(s *state) error {
// 	return nil
// }

// func opSwap3(s *state) error {
// 	return nil
// }

// func opSwap4(s *state) error {
// 	return nil
// }

// func opSwap5(s *state) error {
// 	return nil
// }

// func opSwap6(s *state) error {
// 	return nil
// }

// func opSwap7(s *state) error {
// 	return nil
// }

// func opSwap8(s *state) error {
// 	return nil
// }

// func opSwap9(s *state) error {
// 	return nil
// }

// func opSwap10(s *state) error {
// 	return nil
// }

// func opSwap11(s *state) error {
// 	return nil
// }

// func opSwap12(s *state) error {
// 	return nil
// }

// func opSwap13(s *state) error {
// 	return nil
// }

// func opSwap14(s *state) error {
// 	return nil
// }

// func opSwap15(s *state) error {
// 	return nil
// }

// func opSwap16(s *state) error {
// 	return nil
// }
