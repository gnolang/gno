package evm

import (
	"gno.land/p/demo/uint256"
)

// const memSize = 30000
const gasPrice = 10 // one gas cost 10 wei

type LogEntry struct {
	Topics [][]byte
	Data   []byte
}

type tmpEnv struct {
	storage     map[uint256.Uint]*uint256.Uint
	warmStorage []uint256.Uint
	logs        []LogEntry
}

type State struct {
	caller           *Caller
	callValue        *uint256.Uint
	callData         []byte
	memory           []byte
	result           []byte
	transientStorage map[uint256.Uint]*uint256.Uint
	stack            evmStack
	code             instructions
	tmp              tmpEnv
	gasConsumed      uint // The real amount of gas consumed by the program
	gasBet           int  // The amount of gas specified by user
	env              *Env
}

func newState() *State {
	return &State{
		stack:     *NewevmStack(),
		memory:    []byte{}, // make([]byte, 1024)
		result:    []byte{}, // make([]byte, 1024)
		code:      instructions{},
		caller:    nil,
		callValue: uint256.Zero(),
		tmp: tmpEnv{
			storage:     make(map[uint256.Uint]*uint256.Uint),
			logs:        []LogEntry{},
			warmStorage: make([]uint256.Uint, 0),
		},
		transientStorage: map[uint256.Uint]*uint256.Uint{},
		env:              nil,
		gasConsumed:      0,
		gasBet:           -1,
	}
}

// getters

func (s State) getStackData() []*uint256.Uint {
	return s.stack.data
}

func (s State) getResult() []byte {
	return s.result
}

// setters

func (s *State) setEnv(env *Env) {
	s.env = env
}

func (s *State) setGasBet(gas uint) {
	s.gasBet = int(gas)
}

func (s *State) setCaller(caller *Caller) {
	s.caller = caller
}

func (s *State) setCallValue(wei *uint256.Uint) {
	s.callValue = wei.Clone()
}

func (s *State) setCallData(data []byte) {
	s.callData = append(s.callData, data...)
}

func (s *State) executeByteCode(code []byte) error {
	s.code.Reset(code)
	if s.env != nil {
		s.tmp.storage = make(map[uint256.Uint]*uint256.Uint)
		for k, v := range s.env.Storage {
			s.tmp.storage[k] = v
		}
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
		if s.code.IsDone() { // break before increment to avoid out of bounds
			break
		}
		s.code.Increment()
	}
	if s.code.Ill() {
		return s.code.err
	}
	if s.env != nil {
		s.env.Storage = make(map[uint256.Uint]*uint256.Uint)
		for k, v := range s.tmp.storage {
			s.env.Storage[k] = v
		}
	}
	if s.caller != nil {
		s.caller.Gas -= s.gasConsumed
	}
	return nil
}
