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
)

var validatorPubKeyTypeURLs = map[string]struct{}{
	amino.GetTypeURL(ed25519.PubKeyEd25519{}): {},
}

func DefaultConsensusParams() abci.ConsensusParams {
	return abci.ConsensusParams{
		DefaultBlockParams(),
		DefaultValidatorParams(),
	}
}

func DefaultBlockParams() *abci.BlockParams {
	return &abci.BlockParams{
		MaxTxBytes:   1024 * 1024, // 1MB
		MaxDataBytes: 22020096,    // 21MB
		MaxGas:       -1,
		TimeIotaMS:   1000, // 1s
	}
}

func DefaultValidatorParams() *abci.ValidatorParams {
	return &abci.ValidatorParams{[]string{
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
	for i := 0; i < len(params.Validator.PubKeyTypeURLs); i++ {
		keyType := params.Validator.PubKeyTypeURLs[i]
		if _, ok := validatorPubKeyTypeURLs[keyType]; !ok {
			return errors.New("params.Validator.PubKeyTypeURLs[%d], %s, is an unknown pubKey type",
				i, keyType)
		}
	}

	return nil
}
