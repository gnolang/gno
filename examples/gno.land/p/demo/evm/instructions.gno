package evm

import (
	"errors"
	"strconv"
)

func opNotImplemented(rip *int, codes []byte, stack *ByteStack, memory []byte) error {
	return errors.New("op code " + strconv.FormatInt(int64(codes[*rip]), 16) + "not implemented")
}

func opAdd(rip *int, codes []byte, stack *ByteStack, memory []byte) error {
	a1, err := stack.Pop()
	if err != nil {
		return errors.New("can't ADD: empty stack")
	}

	a2, err := stack.Pop()
	if err != nil {
		return errors.New("can't ADD: stack of size 1")
	}

	stack.Push(a1 + a2)
	*rip += 1
	return nil
}
