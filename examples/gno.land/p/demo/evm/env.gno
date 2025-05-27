package evm

import (
	"gno.land/p/demo/uint256"
)

// Env is a global environnement to evaluate your EVM state
type Env struct {
	Logs            []LogEntry
	Storage         map[uint256.Uint]*uint256.Uint
	Accounts        Accounts
	NextAccountId   uint64
	Transaction     Transactions
	ChainID         *uint256.Uint
	BlockDifficulty *uint256.Uint
}

func newEnv() *Env {
	return &Env{
		Logs: []LogEntry{},
		Storage: map[uint256.Uint]*uint256.Uint{},
		Accounts: Accounts{},
		NextAccountId: 0x10,
		Transaction: Transactions{},
		ChainID: uint256.Zero(),
		BlockDifficulty: uint256.Zero(),
	}
}
