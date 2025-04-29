package evm

import (
	"gno.land/p/demo/uint256"
)

// const memSize = 30000

type LogEntry struct {
	Topics [][]byte
	Data   []byte
}

type tmpEnv struct {
	storage []byte
	logs    []LogEntry
}

type State struct {
	memory           []byte
	result 			 []byte
	transientStorage map[uint256.Uint]uint256.Uint
	stack            evmStack
	code             instructions
	tmp              tmpEnv
	gasConsumed      uint		// The real amount of gas consumed by the program
	gasBet           int 		// The amount of gas specified by user
	env              *Env
}

func newState() *State {
	return &State{
		stack:  *NewevmStack(),
		memory: []byte{}, // make([]byte, 1024)
		result: []byte{}, // make([]byte, 1024)
		code:   instructions{},
		tmp: tmpEnv{
			storage: []byte{},
			logs:    []LogEntry{},
		},
		transientStorage: map[uint256.Uint]uint256.Uint{},
		env:              nil,
		gasConsumed:      0,
		gasBet:           -1,
	}
}

// TODO: return the s.buf
func (s State) getStackData() []*uint256.Uint {
	// return s.buf
	return s.stack.data
}

func (s State) getResult() []byte {
	return s.result
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
	for !s.code.IsDone() && s.code.Current() != byte(STOP) {
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
	// TODO: substract the gas amount
	return nil
}
