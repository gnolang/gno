package evm

// Env is a global environnement to evaluate your EVM state
type Env struct {
	Logs []byte
	Storage []byte
}

func (e *Env) Eval(st *State) {

}

func (e *Env) Execute(code []byte) {
	// st := newState()
	// st.executeByteCode()
}