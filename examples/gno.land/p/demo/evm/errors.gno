package evm

import (
	"errors"
	"strconv"
)

var (
	outOfBounds = errors.New("out of bounds")
	invalidInstruction = errors.New("invalid instruction")
	outOfGas = errors.New("the user is out of gas")
)

func notImplemented(s *State) error {
	return errors.New("op code 0x" + strconv.FormatInt(int64(s.code.Current()), 16) + " not implemented")
}