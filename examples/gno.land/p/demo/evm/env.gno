package evm

import (
	"gno.land/p/demo/uint256"
)

const nbPrecompiledContract = 0x0a

// Env is a global environnement to evaluate your EVM state
type Env struct {
	Logs            []LogEntry
	Storage         map[uint256.Uint]*uint256.Uint
	Accounts        Accounts
	NextAccountId   uint64
	Transaction     Transactions
	ChainID         *uint256.Uint
	BlockDifficulty *uint256.Uint
	BlobBaseFee		*uint256.Uint
}

func NewEnv() *Env {
	return &Env{
		Logs:            []LogEntry{},
		Storage:         map[uint256.Uint]*uint256.Uint{},
		Accounts:        Accounts{},
		NextAccountId:   nbPrecompiledContract + 1,
		Transaction:     Transactions{},
		ChainID:         uint256.Zero(),
		BlockDifficulty: uint256.Zero(),
		BlobBaseFee: 	 uint256.Zero(),
	}
}
