package evm

import "gno.land/p/demo/uint256"

// Env is a global environnement to evaluate your EVM state
type Env struct {
	Logs    []byte
	Storage map[uint256.Uint]*uint256.Uint
}
