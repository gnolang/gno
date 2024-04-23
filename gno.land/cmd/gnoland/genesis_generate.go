package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var defaultChainID = "dev"

type generateCfg struct {
	outputPath        string
	chainID           string
	genesisTime       int64
	blockMaxTxBytes   int64
	blockMaxDataBytes int64
	blockMaxGas       int64
	blockTimeIota     int64
}

// newGenerateCmd creates the genesis generate subcommand
func newGenerateCmd(io commands.IO) *commands.Command {
	cfg := &generateCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "generate",
			ShortUsage: "generate [flags]",
			ShortHelp:  "generates a fresh genesis.json",
			LongHelp:   "Generates a node's genesis.json based on specified parameters",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execGenerate(cfg, io)
		},
	)
}

func (c *generateCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.outputPath,
		"output-path",
		"./genesis.json",
		"the output path for the genesis.json",
	)

	fs.Int64Var(
		&c.genesisTime,
		"genesis-time",
		time.Now().Unix(),
		"the genesis creation time. Defaults to current time",
	)

	fs.StringVar(
		&c.chainID,
		"chain-id",
		defaultChainID,
		"the ID of the chain",
	)

	fs.Int64Var(
		&c.blockMaxTxBytes,
		"block-max-tx-bytes",
		types.MaxBlockTxBytes,
		"the max size of the block transaction",
	)

	fs.Int64Var(
		&c.blockMaxDataBytes,
		"block-max-data-bytes",
		types.MaxBlockDataBytes,
		"the max size of the block data",
	)

	fs.Int64Var(
		&c.blockMaxGas,
		"block-max-gas",
		types.MaxBlockMaxGas,
		"the max gas limit for the block",
	)

	fs.Int64Var(
		&c.blockTimeIota,
		"block-time-iota",
		types.BlockTimeIotaMS,
		"the block time iota (in ms)",
	)
}

func execGenerate(cfg *generateCfg, io commands.IO) error {
	// Start with the default configuration
	genesis := getDefaultGenesis()

	// Set the genesis time
	if cfg.genesisTime > 0 {
		genesis.GenesisTime = time.Unix(cfg.genesisTime, 0)
	}

	// Set the chain ID
	if cfg.chainID != "" {
		genesis.ChainID = cfg.chainID
	}

	// Set the max tx bytes
	if cfg.blockMaxTxBytes > 0 {
		genesis.ConsensusParams.Block.MaxTxBytes = cfg.blockMaxTxBytes
	}

	// Set the max data bytes
	if cfg.blockMaxDataBytes > 0 {
		genesis.ConsensusParams.Block.MaxDataBytes = cfg.blockMaxDataBytes
	}

	// Set the max block gas
	if cfg.blockMaxGas > 0 {
		genesis.ConsensusParams.Block.MaxGas = cfg.blockMaxGas
	}

	// Set the block time IOTA
	if cfg.blockTimeIota > 0 {
		genesis.ConsensusParams.Block.TimeIotaMS = cfg.blockTimeIota
	}

	// Validate the genesis
	if validateErr := genesis.ValidateAndComplete(); validateErr != nil {
		return fmt.Errorf("unable to validate genesis, %w", validateErr)
	}

	// Save the genesis file to disk
	if saveErr := genesis.SaveAs(cfg.outputPath); saveErr != nil {
		return fmt.Errorf("unable to save genesis, %w", saveErr)
	}

	io.Printfln("Genesis successfully generated at %s\n", cfg.outputPath)

	// Log the empty validator set warning
	io.Printfln("WARN: Genesis is generated with an empty validator set")

	return nil
}

// getDefaultGenesis returns the default genesis config
func getDefaultGenesis() *types.GenesisDoc {
	return &types.GenesisDoc{
		GenesisTime:     time.Now(),
		ChainID:         defaultChainID,
		ConsensusParams: types.DefaultConsensusParams(),
	}
}
