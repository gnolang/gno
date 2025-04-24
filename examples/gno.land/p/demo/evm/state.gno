package evm

import (
	"gno.land/p/demo/uint256"
)

// const memSize = 30000

type state struct {
	memory []byte
	stack  evmStack
	buf    []byte
	code   instructions
}

func newState() *state {
	return &state{
		stack:  *NewevmStack(),
		memory: []byte{},
		buf:    []byte{},
		code:   instructions{},
	}
}

// TODO: return the s.buf
func (s state) getStackData() []*uint256.Uint {
	// return s.buf
	return s.stack.data
}

func (s state) getResult() []byte {
	return s.buf
}

func (s *state) executeByteCode(code []byte) error {
	s.code.Reset(code)
	for !s.code.Ill() && s.code.Current() != byte(STOP) {
		opData, ok := opCodes[OpCode(s.code.Current())]
		if !ok {
			return opNotImplemented(s)
		}
		err := opData.operation(s)
		if err != nil {
			return err
		}
	}
	if s.code.Ill() {
		return outOfBounds
	}
	return nil
}
