package evm

import (
	"gno.land/p/demo/uint256"
	"fmt"
)

// const memSize = 30000

type LogEntry struct {
    Topics [][]byte
    Data   []byte
}

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
	gasBet      int
	env         *Env
	logs   		[]LogEntry
}

func newState() *State {
	return &State{
		stack:  *NewevmStack(),
		memory: []byte{}, // make([]byte, 1024)
		code:   instructions{},
		tmp: tmpEnv{
			storage: []byte{},
			logs:    []byte{},
		},
		env:         nil,
		gasConsumed: 0,
		gasBet:      -1,
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

func (s *State) setEnv(env *Env) {
	s.env = env
}

func (s *State) setGasBet(gas uint) {
	s.gasBet = int(gas)
}

func (s *State) executeByteCode(code []byte) error {
	s.code.Reset(code)
	if s.env != nil {
		s.tmp.storage = make([]byte, len(s.env.Storage))
		copy(s.tmp.storage, s.env.Storage)
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
		if s.gasBet > 0 && s.gasBet < int(s.gasConsumed) {
			return outOfGas
		}
		s.code.Increment()
	}
	if s.code.Ill() {
		return s.code.err
	}
	if s.env != nil {
		copy(s.env.Storage, s.tmp.storage)
	}
	return nil
}
