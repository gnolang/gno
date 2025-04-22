package evm

const memSize = 30000

type State struct {
	memory []byte
	stack  ByteStack
	buf    []byte
	code   Instructions
}

func newState() *State {
	return &State{
		stack:  *NewByteStack(),
		memory: make([]byte, memSize),
		buf:    []byte{},
		code:   Instructions{},
	}
}

// TODO: return the s.buf
func (s State) getResult() []byte {
	// return s.buf
	return s.stack.data
}

func (s *State) executeByteCode(code []byte) error {
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
