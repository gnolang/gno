package types

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
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

	// MaxBlockMaxGas is the max gas limit for the block
	MaxBlockMaxGas int64 = 3000000000 // 3B gas

	// BlockTimeIotaMS is the block time iota (in ms)
	BlockTimeIotaMS int64 = 100 // ms
)

var validatorPubKeyTypeURLs = map[string]struct{}{
	amino.GetTypeURL(ed25519.PubKeyEd25519{}):     {},
	amino.GetTypeURL(secp256k1.PubKeySecp256k1{}): {},
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
		amino.GetTypeURL(secp256k1.PubKeySecp256k1{}),
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
