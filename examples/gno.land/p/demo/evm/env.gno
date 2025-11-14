package evm

import (
	"gno.land/p/onbloc/uint256"
)

const nbPrecompiledContract = 0x0a

// Env is a global environnement to evaluate your EVM state
type Env struct {
	Logs            []logEntry               // Logs emitted during execution (for events)
	Accounts        Accounts                 // List of accounts (EOA and contracts)
	NextAccountId   uint64                   // Next account ID for contract creation
	Coinbase        *uint256.Uint            // Miner address
	Transaction     Transactions             // Current transaction context
	ChainID         *uint256.Uint            // Chain ID for replay protection (EIP-155)
	BlockDifficulty *uint256.Uint            // Block difficulty (used in PoW, ignored in PoS)
	BlockNumber     *uint256.Uint            // Current block number
	BlockHashes     map[uint64]*uint256.Uint // Map block number -> block hash
	BlobHashes      []*uint256.Uint          // Hashes of blob data (EIP-4844)
	BlobBaseFee     *uint256.Uint            // Base fee for blob transactions (EIP-4844)
	PrevBaseFee     uint64                   // Previous block's base fee (EIP-1559)
	IncludedGas     uint64                   // Gas used in the current block (EIP-1559)
	TargetGas       uint64                   // Target gas limit per block (EIP-1559)
}

// Create a new environnement, with default values.
func NewEnv() *Env {
	return &Env{
		Logs:            []logEntry{},
		Accounts:        Accounts{},
		NextAccountId:   nbPrecompiledContract + 1,
		Coinbase:        uint256.Zero(),
		Transaction:     Transactions{},
		ChainID:         uint256.Zero(),
		BlockDifficulty: uint256.Zero(),
		BlockNumber:     uint256.Zero(),
		BlockHashes:     make(map[uint64]*uint256.Uint),
		BlobHashes:      []*uint256.Uint{},
		BlobBaseFee:     uint256.Zero(),
		PrevBaseFee:     0,
		IncludedGas:     0,
		TargetGas:       0,
	}
}
