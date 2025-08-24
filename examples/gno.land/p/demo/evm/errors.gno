package evm

import (
	"errors"
	"strconv"
)

var (
	outOfBounds        = errors.New("out of bounds")
	invalidInstruction = errors.New("invalid instruction")
	outOfGas           = errors.New("the user is out of gas")
	notEnoughGas       = errors.New("the user does not have that much gas")
	notEnoughWei       = errors.New("the user does not have that much wei")
	tooFewStack        = errors.New("too few values on stack")
	invalidPanic       = errors.New("invalid was called, aborting")
	revertError        = errors.New("reverted")
)

func notImplemented(s *state) error {
	return errors.New("op code 0x" + strconv.FormatInt(int64(s.code.current()), 16) + " not implemented")
}
