package evm

import (
	"gno.land/p/demo/evm/uint256"
)

const memSize = 30000

type state struct {
	memory []byte
	stack  evmStack
	buf    []byte
	code   instructions
}

func newState() *state {
	return &state{
		stack:  *NewevmStack(),
		memory: make([]byte, memSize),
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
	for !s.code.Ill() && s.code.Current() != byte(STOP) { // 60 01 60 02 01 00
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
