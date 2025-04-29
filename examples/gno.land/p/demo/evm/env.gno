package evm

// Env is a global environnement to evaluate your EVM state
type Env struct {
	Logs []byte
	Storage []byte
}
