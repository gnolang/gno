package types

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

const (
	// MaxBlockSizeBytes is the maximum permitted size of the blocks.
	MaxBlockSizeBytes = 104857600 // 100MB

	// BlockPartSizeBytes is the size of one block part.
	BlockPartSizeBytes = 65536 // 64kB

	// MaxBlockPartsCount is the maximum count of block parts.
	MaxBlockPartsCount = (MaxBlockSizeBytes / BlockPartSizeBytes) + 1

	// MaxBlockTxBytes is the max size of the block transaction
	MaxBlockTxBytes int64 = 1000000 // 1MB

	// MaxBlockDataBytes is the max size of the block data
	MaxBlockDataBytes int64 = 2000000 // 2MB

	// MaxBlockDataBytesLimit is the hard upper bound on the configurable
	// Block.MaxDataBytes consensus parameter. Together with a fixed allowance
	// for the block header and commit it bounds the total serialized size of a
	// block, which in turn bounds the fast-sync block-response message (see
	// maxMsgSize in the blockchain reactor). Raising MaxDataBytes above this
	// would let the chain produce blocks that fast-syncing peers reject,
	// stalling sync, so it is rejected at consensus-param validation.
	MaxBlockDataBytesLimit int64 = 8 << 20 // 8MB

	// MaxBlockOverheadBytes is the allowance that Block.MaxTxBytes has to leave
	// inside Block.MaxDataBytes for everything in a serialized block that is not
	// tx data: the header, the LastCommit, and amino's framing. Measured, a
	// block costs 428 bytes empty plus ~167 bytes per validator in its
	// LastCommit, plus 44 bytes of framing per tx, so 128KB covers a commit for
	// roughly 780 validators.
	//
	// Without the room, a single tx whose raw size fits MaxTxBytes produces a
	// block that does not fit MaxDataBytes -- which is also the size peers
	// decode a proposal with. Such a tx is admitted by CheckTx, reaped on its
	// own (ReapMaxBytesMaxGas stops at the first tx that does not fit rather
	// than skipping it), and then trimmed straight back out by
	// CreateProposalBlock, so it is never committed, never evicted, and every tx
	// queued behind it starves.
	MaxBlockOverheadBytes int64 = 128 << 10 // 128KB

	// MaxBlockMaxGas is the max gas limit for the block
	MaxBlockMaxGas int64 = 3000000000 // 3B gas

	// BlockTimeIotaMS is the block time iota (in ms)
	BlockTimeIotaMS int64 = 100 // ms
)

var validatorPubKeyTypeURLs = map[string]struct{}{
	amino.GetTypeURL(ed25519.PubKeyEd25519{}): {},
}

func DefaultConsensusParams() abci.ConsensusParams {
	return abci.ConsensusParams{
		Block:     DefaultBlockParams(),
		Validator: DefaultValidatorParams(),
	}
}

func DefaultBlockParams() *abci.BlockParams {
	return &abci.BlockParams{
		MaxTxBytes:   MaxBlockTxBytes,
		MaxDataBytes: MaxBlockDataBytes,
		MaxGas:       MaxBlockMaxGas,
		TimeIotaMS:   BlockTimeIotaMS,
	}
}

func DefaultValidatorParams() *abci.ValidatorParams {
	return &abci.ValidatorParams{PubKeyTypeURLs: []string{
		amino.GetTypeURL(ed25519.PubKeyEd25519{}),
	}}
}

func ValidateConsensusParams(params abci.ConsensusParams) error {
	if params.Block.MaxTxBytes <= 0 {
		return errors.New("Block.MaxTxBytes must be greater than 0. Got %d",
			params.Block.MaxTxBytes)
	}
	if params.Block.MaxTxBytes > MaxBlockSizeBytes {
		return errors.New("Block.MaxTxBytes is too big. %d > %d",
			params.Block.MaxTxBytes, MaxBlockSizeBytes)
	}

	// A non-positive MaxDataBytes is not a benign "unlimited": 0 panics the
	// proposer in mempool.ReapMaxBytesMaxGas, and a negative value both disables
	// the reaping limit (so nothing bounds a block's data size, defeating the
	// MaxBlockDataBytesLimit ceiling below) and panics amino when the consensus
	// state decodes a proposal block with it as the max size.
	if params.Block.MaxDataBytes <= 0 {
		return errors.New("Block.MaxDataBytes must be greater than 0. Got %d",
			params.Block.MaxDataBytes)
	}
	if params.Block.MaxDataBytes > MaxBlockDataBytesLimit {
		return errors.New("Block.MaxDataBytes is too big. %d > %d",
			params.Block.MaxDataBytes, MaxBlockDataBytesLimit)
	}

	// MaxDataBytes bounds the whole serialized block, not just its tx data, so a
	// single MaxTxBytes-sized tx has to leave room for the header and commit.
	// See MaxBlockOverheadBytes for what goes wrong when it does not.
	if params.Block.MaxTxBytes+MaxBlockOverheadBytes > params.Block.MaxDataBytes {
		return errors.New("Block.MaxTxBytes must leave %d bytes of Block.MaxDataBytes for the block header and commit. %d + %d > %d",
			MaxBlockOverheadBytes, params.Block.MaxTxBytes, MaxBlockOverheadBytes, params.Block.MaxDataBytes)
	}

	if params.Block.MaxGas < -1 {
		return errors.New("Block.MaxGas must be greater or equal to -1. Got %d",
			params.Block.MaxGas)
	}

	if params.Block.TimeIotaMS <= 0 {
		return errors.New("Block.TimeIotaMS must be greater than 0. Got %v",
			params.Block.TimeIotaMS)
	}

	if len(params.Validator.PubKeyTypeURLs) == 0 {
		return errors.New("len(Validator.PubKeyTypeURLs) must be greater than 0")
	}

	// Check if keyType is a known ABCIPubKeyType
	for i := range params.Validator.PubKeyTypeURLs {
		keyType := params.Validator.PubKeyTypeURLs[i]
		if _, ok := validatorPubKeyTypeURLs[keyType]; !ok {
			return errors.New("params.Validator.PubKeyTypeURLs[%d], %s, is an unknown pubKey type",
				i, keyType)
		}
	}

	return nil
}
