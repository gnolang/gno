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
	accounts    Accounts
}

type State struct {
	callerAddress    uint64
	callValue        *uint256.Uint
	callData         []byte
	memory           []byte
	result           []byte
	transientStorage map[uint256.Uint]*uint256.Uint
	stack            evmStack
	code             instructions
	tmp              tmpEnv
	gasConsumed      uint // The real amount of gas consumed by the program
	gasLimit         int  // The amount of gas specified by user
	lastCallState    *State
	env              *Env
}

func newEmptyState() *State {
	return &State{
		stack:         *NewevmStack(),
		memory:        []byte{}, // make([]byte, 1024)
		result:        []byte{}, // make([]byte, 1024)
		code:          instructions{},
		callerAddress: 0,
		callValue:     uint256.Zero(),
		tmp: tmpEnv{
			storage:     map[uint256.Uint]*uint256.Uint{},
			logs:        []LogEntry{},
			warmStorage: []uint256.Uint{},
			accounts:    Accounts{},
		},
		transientStorage: map[uint256.Uint]*uint256.Uint{},
		env:              nil,
		gasConsumed:      0,
		gasLimit:         -1,
	}
}
func newStateFromEnv(env *Env) *State {
	storageCopy := make(map[uint256.Uint]*uint256.Uint)
	if len(env.Storage) > 0 {
		for k, v := range env.Storage {
			storageCopy[k] = v
		}
	}

	logsCopy := make([]LogEntry, len(env.Logs))
	copy(logsCopy, env.Logs)

	accountsCopy := make(Accounts, 0)
	if len(env.Accounts) > 0 {
		for k, v := range env.Accounts {
			accountsCopy[k] = v
		}
	}
	return &State{
		stack:         *NewevmStack(),
		memory:        []byte{}, // make([]byte, 1024)
		result:        []byte{}, // make([]byte, 1024)
		code:          instructions{},
		callerAddress: 0,
		callValue:     uint256.Zero(),
		tmp: tmpEnv{
			storage:     storageCopy,
			logs:        logsCopy,
			warmStorage: []uint256.Uint{},
			accounts:    accountsCopy,
		},
		transientStorage: map[uint256.Uint]*uint256.Uint{},
		env:              env,
		gasConsumed:      0,
		gasLimit:         -1,
	}
}

func (s *State) clone() *State {
	storageCopy := make(map[uint256.Uint]*uint256.Uint)
	for k, v := range s.tmp.storage {
		storageCopy[k] = v
	}

	logsCopy := make([]LogEntry, len(s.tmp.logs))
	copy(logsCopy, s.tmp.logs)

	warmStorageCopy := make([]uint256.Uint, len(s.tmp.warmStorage))
	copy(warmStorageCopy, s.tmp.warmStorage)

	accountsCopy := make(Accounts)
	for k, v := range s.tmp.accounts {
		accountsCopy[k] = v
	}

	return &State{
		stack:         *NewevmStack(),
		memory:        []byte{},
		result:        []byte{},
		code:          instructions{},
		callerAddress: s.callerAddress,
		callValue:     uint256.Zero(),
		tmp: tmpEnv{
			storage:     storageCopy,
			logs:        logsCopy,
			warmStorage: warmStorageCopy,
			accounts:    accountsCopy,
		},
		transientStorage: map[uint256.Uint]*uint256.Uint{},
		env:              s.env,
		gasConsumed:      0,
		gasLimit:         -1,
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
func (s *State) setCode(code []byte) {
	s.code.Reset(code)
}

func (s *State) setEnv(env *Env) {
	s.env = env
}

func (s *State) setGasLimit(gas uint) {
	s.gasLimit = int(gas)
}

func (s *State) setCallerAddress(callerAddress uint64) {
	s.callerAddress = callerAddress
}

func (s *State) setCallValue(wei *uint256.Uint) {
	s.callValue = wei.Clone()
}

func (s *State) setCallData(data []byte) {
	s.callData = append(s.callData, data...)
}

// usefull
func (s *State) loadStagedChangesIntoEnv(env *Env) {
	env.Storage = make(map[uint256.Uint]*uint256.Uint)
	for k, v := range s.tmp.storage {
		env.Storage[k] = v
	}

	env.Logs = make([]LogEntry, len(s.tmp.logs))
	copy(env.Logs, s.tmp.logs)

	env.Accounts = make(Accounts)
	for k, v := range s.tmp.accounts {
		env.Accounts[k] = v
	}
}

func (s *State) executeByteCode() error {
	for !s.code.EndOfCode() && s.code.Current() != byte(STOP) {
		opData, ok := opCodes[OpCode(s.code.Current())]
		if !ok {
			return invalidInstruction
		}
		s.gasConsumed += uint(opData.gas)
		err := opData.operation(s)
		if err != nil {
			return err
		}
		if s.gasLimit > 0 && s.gasLimit < int(s.gasConsumed) {
			return outOfGas
		}
		if s.code.IsDone() { // break before increment to avoid out of bounds
			break
		}
		if s.code.Ill() {
			return s.code.err
		}
		s.code.Increment()
	}
	return nil
}
