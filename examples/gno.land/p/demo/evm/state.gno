package evm

import (
	"gno.land/p/demo/uint256"
)

// const memSize = 30000

type tmpEnv struct {
	storage []byte
	logs    []byte
}

type State struct {
	memory      []byte
	stack       evmStack
	code        instructions
	tmp         tmpEnv
	gasConsumed uint
	parent      *Env
}

func newState() *State {
	return &State{
		stack:  *NewevmStack(),
		memory: []byte{},
		code:   instructions{},
		tmp: tmpEnv{
			storage: []byte{},
			logs:    []byte{},
		},
		parent: nil,
		gasConsumed: 0,
	}
}

// TODO: return the s.buf
func (s State) getStackData() []*uint256.Uint {
	// return s.buf
	return s.stack.data
}

func (s State) getResult() []byte {
	return []byte{}
}

func (s *State) openEnv(env *Env) {
	s.parent = env
}

func (s *State) executeByteCodeWithGas(code []byte, gas int) error {
	s.code.Reset(code)
	if s.parent != nil {
		s.tmp.storage = make([]byte, len(s.parent.Storage))
		copy(s.tmp.storage, s.parent.Storage)
	}
	for !s.code.Ill() && s.code.Current() != byte(STOP) {
		opData, ok := opCodes[OpCode(s.code.Current())]
		if !ok {
			return invalidInstruction
		}
		s.gasConsumed += uint(opData.gas)
		err := opData.operation(s)
		if err != nil {
			return err
		}
		if gas > 0 && gas < int(s.gasConsumed) {
			return outOfGas
		}
		s.code.Increment()
	}
	if s.code.Ill() {
		return s.code.err
	}
	if s.parent != nil {
		copy(s.parent.Storage, s.tmp.storage)
	}
	return nil
}
