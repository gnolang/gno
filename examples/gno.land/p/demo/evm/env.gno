package evm

import (
	"gno.land/p/demo/uint256"
)

// Env is a global environnement to evaluate your EVM state
type Env struct {
	Logs            []LogEntry
	Storage         map[uint256.Uint]*uint256.Uint
	Accounts        Accounts
	NextAccountId   uint256.Uint
	Transaction     Transactions
	ChainID         *uint256.Uint
	BlockDifficulty *uint256.Uint
}
