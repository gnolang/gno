package evm

import (
	"gno.land/p/demo/uint256"
)

type contextType uint

// gasPrice defines the cost of one gas unit in wei.
const (
	GasPrice    uint64      = 10
	callContext contextType = iota
	staticContext
) // one gas cost 10 wei

// LogEntry represents an event log entry with topics and data.
type LogEntry struct {
	Topics [][]byte // Indexed topics for filtering logs
	Data   []byte   // Raw data associated with the log entry
}

// tmpEnv stores temporary execution environment data such as storage, logs, warm storage slots, and accounts.
type tmpEnv struct {
	warmStorage        []uint256.Uint // List of storage slots accessed ("warm" slots)
	warmAccountAddress []uint256.Uint // List of account addresses accessed ("warm" address)
	logs               []LogEntry     // Event logs generated during execution
	accounts           Accounts       // Temporary accounts state
	transactions       Transactions   // Temporary list of transactions that happened on the chain
}

// State represents the full EVM execution state during a transaction or message call.
type State struct {
	context          contextType
	originAddress    uint64                         // Address of the caller that inititated the transaction
	callerAddress    uint64                         // Address of the caller executing this state
	callValue        *uint256.Uint                  // Amount of wei sent with the call
	callData         []byte                         // Input data for the call
	memory           []byte                         // Linear byte-addressable memory of the EVM
	result           []byte                         // Output data from the execution
	transientStorage map[uint256.Uint]*uint256.Uint // Transient (temporary) storage key-value pairs (not persisted)
	stack            evmStack                       // EVM stack structure
	code             instructions                   // Bytecode instructions being executed
	tmp              tmpEnv                         // Temporary execution environment data
	gasConsumed      uint                           // Total gas consumed so far during execution
	gasLimit         int                            // Maximum gas allowed for execution
	lastCallState    *State                         // Pointer to the previous call's state, if any
	env              *Env                           // Reference to the external environment contextF
}

// newEmptyState initializes and returns a new State with empty/default values.
func newEmptyState() *State {
	return &State{
		context:       callContext,
		stack:         *NewevmStack(),
		memory:        []byte{}, // make([]byte, 1024)
		result:        []byte{}, // make([]byte, 1024)
		code:          instructions{},
		callerAddress: 0,
		callValue:     uint256.Zero(),
		tmp: tmpEnv{
			logs:         []LogEntry{},
			warmStorage:  []uint256.Uint{},
			accounts:     Accounts{},
			transactions: []Transaction{},
		},
		transientStorage: map[uint256.Uint]*uint256.Uint{},
		env:              nil,
		gasConsumed:      0,
		gasLimit:         -1,
	}
}

// newStateFromEnv creates a new State initialized from a given external Env, copying storage, logs, and accounts.
func newStateFromEnv(env *Env) *State {

	logsCopy := make([]LogEntry, len(env.Logs))
	copy(logsCopy, env.Logs)

	transactionsCopy := make(Transactions, len(env.Transaction))
	copy(transactionsCopy, env.Transaction)

	return &State{
		context:       callContext,
		stack:         *NewevmStack(),
		memory:        []byte{}, // make([]byte, 1024)
		result:        []byte{}, // make([]byte, 1024)
		code:          instructions{},
		callerAddress: 0,
		callValue:     uint256.Zero(),
		tmp: tmpEnv{
			logs:         logsCopy,
			warmStorage:  []uint256.Uint{},
			accounts:     *env.Accounts.Clone(),
			transactions: transactionsCopy,
		},
		transientStorage: map[uint256.Uint]*uint256.Uint{},
		env:              env,
		gasConsumed:      0,
		gasLimit:         -1,
	}
}

// clone creates a deep copy of the current State, including storage, logs, warm storage, and accounts.
// Note that stack, memory, and code are reset/empty in the clone.
func (s *State) clone() *State {
	logsCopy := make([]LogEntry, len(s.tmp.logs))
	copy(logsCopy, s.tmp.logs)

	warmStorageCopy := make([]uint256.Uint, len(s.tmp.warmStorage))
	copy(warmStorageCopy, s.tmp.warmStorage)

	transactionsCopy := make(Transactions, len(s.tmp.transactions))
	copy(transactionsCopy, s.tmp.transactions)

	return &State{
		context:       s.context,
		stack:         *NewevmStack(),
		memory:        []byte{},
		result:        []byte{},
		code:          instructions{},
		callerAddress: s.callerAddress,
		originAddress: s.originAddress,
		callValue:     s.callValue.Clone(),
		tmp: tmpEnv{
			logs:         logsCopy,
			warmStorage:  warmStorageCopy,
			accounts:     *s.tmp.accounts.Clone(),
			transactions: transactionsCopy,
		},
		transientStorage: map[uint256.Uint]*uint256.Uint{},
		env:              s.env,
		gasConsumed:      0,
		gasLimit:         -1,
	}
}

// reapply a state at the end of a sub context
func (s *State) applyStateAfterReturn(tmpState *State) {
	logsCopy := make([]LogEntry, len(tmpState.tmp.logs))
	copy(logsCopy, tmpState.tmp.logs)

	warmStorageCopy := make([]uint256.Uint, len(tmpState.tmp.warmStorage))
	copy(warmStorageCopy, tmpState.tmp.warmStorage)

	transactionsCopy := make(Transactions, len(tmpState.tmp.transactions))
	copy(transactionsCopy, s.tmp.transactions)

	s.lastCallState = tmpState
	s.tmp = tmpEnv{
		logs:         logsCopy,
		warmStorage:  warmStorageCopy,
		accounts:     *tmpState.tmp.accounts.Clone(),
		transactions: transactionsCopy,
	}
}

//==// GETTERS //==//

// getStackData returns a slice of pointers to the current data on the stack.
func (s State) getStackData() []*uint256.Uint {
	return s.stack.data
}

// getResult returns the result (output) byte slice of the current execution.
func (s State) getResult() []byte {
	return s.result
}

//==// SETTERS //==//

// setCode sets the contract bytecode to be executed by resetting the code instructions.
func (s *State) setCode(code []byte) {
	s.code.reset(code)
}

// setEnv sets the external environment reference for this State.
func (s *State) setEnv(env *Env) {
	s.env = env
}

// setGasLimit sets the maximum gas limit for this execution State.
func (s *State) setGasLimit(gas int) {
	s.gasLimit = gas
}

// setCallerAddress sets the address of the caller initiating the execution.
func (s *State) setCallerAddress(callerAddress uint64) {
	s.callerAddress = callerAddress
}

// setOriginAddress sets the address of the caller initiating the execution.
func (s *State) setOriginAddress(callerAddress uint64) {
	s.originAddress = callerAddress
}

// setCallValue sets the amount of wei sent with the call.
func (s *State) setCallValue(wei *uint256.Uint) {
	s.callValue = wei.Clone()
}

// setCallData appends input call data bytes to the existing call data.
func (s *State) setCallData(data []byte) {
	s.callData = make([]byte, len(data))
	copy(s.callData, data)

}

//==// USEFULL //==//

// loadStagedChangesIntoEnv copies all staged storage, logs, and accounts changes from the temporary execution environment back into the external Env.
func (s *State) loadStagedChangesIntoEnv(env *Env) {
	env.Logs = make([]LogEntry, len(s.tmp.logs))
	copy(env.Logs, s.tmp.logs)

	env.Accounts = *s.tmp.accounts.Clone()
}

// executeByteCode runs the contract bytecode until a STOP opcode or end of code is reached.
// It updates gas consumption and checks for out-of-gas or invalid instruction errors.
func (s *State) executeByteCode() error {
	for !s.code.endOfCode() && s.code.current() != byte(STOP) {
		opData, ok := opCodes[OpCode(s.code.current())]
		if !ok {
			return invalidInstruction
		}
		s.gasConsumed += opData.gas
		err := opData.operation(s)
		if err != nil {
			return err
		}
		if s.gasLimit >= 0 && s.gasLimit < int(s.gasConsumed) {
			return outOfGas
		}
		if s.code.isDone() { // break before increment to avoid out of bounds
			break
		}
		if s.code.ill() {
			return s.code.err
		}
		s.code.increment()
	}
	return nil
}
