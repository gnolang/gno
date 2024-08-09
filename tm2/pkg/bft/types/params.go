package types

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
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
	MaxBlockMaxGas int64 = 100000000 // 100M gas

	// BlockTimeIotaMS is the block time iota (in ms)
	BlockTimeIotaMS int64 = 100 // ms

	// PriceChangeCompressor is to reduce the impact of gas changes to gas price.
	PriceChangeCompressor int64 = 8 // 1/8 = 0.125

	// BlockTimeIotaMS is the block time iota (in ms)
	TargetGas int64 = 60000000 // 60M gas

	// Protocol level InitialGasPrice  10token/100gas
	// XXX: use std.GasPrice?
	InitialGasPriceAmount int64  = 10
	InitialGasPriceDenom  string = "token"
	InitialGasPriceGas    int64  = 100
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
		MaxTxBytes:            MaxBlockTxBytes,
		MaxDataBytes:          MaxBlockDataBytes,
		MaxGas:                MaxBlockMaxGas,
		TimeIotaMS:            BlockTimeIotaMS,
		PriceChangeCompressor: PriceChangeCompressor,
		TargetGas:             TargetGas,
		InitialGasPriceAmount: InitialGasPriceAmount,
		InitialGasPriceDenom:  InitialGasPriceDenom,
		InitialGasPriceGas:    InitialGasPriceGas,
	}
}

func DefaultValidatorParams() *abci.ValidatorParams {
	return &abci.ValidatorParams{
		PubKeyTypeURLs: []string{
			amino.GetTypeURL(ed25519.PubKeyEd25519{}),
		},
	}
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

	if params.Block.PriceChangeCompressor <= 0 {
		return errors.New("Block.PriceChangeCompressor must be greater than 0. Got %v",
			params.Block.PriceChangeCompressor)
	}

	if params.Block.TargetGas < 0 || params.Block.TargetGas > params.Block.MaxGas {
		return errors.New("Block.TargetGas %v must be greater or equal 0 and smaller than MaxGas %v",
			params.Block.TargetGas, params.Block.MaxGas)
	}

	if params.Block.InitialGasPriceGas <= 0 {
		return errors.New("Block.InitialGasPriceGas must be greater than 0. Got %v",
			params.Block.InitialGasPriceGas)
	}
	if params.Block.InitialGasPriceAmount < 0 {
		return errors.New("Block.InitialGasPriceAmount must be greater or equal to 0. Got %v",
			params.Block.InitialGasPriceAmount)
	}
	err := std.ValidateDenom(params.Block.InitialGasPriceDenom)
	if err != nil {
		return errors.New("Block.InitialGasPriceDenom is not valid, %v", err)
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
