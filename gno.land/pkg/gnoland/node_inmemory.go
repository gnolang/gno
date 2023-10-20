package gnoland

import (
	"fmt"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	tmcfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type InMemoryNodeConfig struct {
	TMConfig              *tmcfg.Config
	ConsensusParams       abci.ConsensusParams
	GenesisValidator      []bft.GenesisValidator
	Packages              []PackagePath
	Balances              []Balance
	GenesisTXs            []std.Tx
	SkipFailingGenesisTxs bool
	GenesisMaxVMCycles    int64
}

// NewInMemoryNodeConfig creates a default configuration for an in-memory node.
func NewInMemoryNodeConfig(tmcfg *tmcfg.Config) *InMemoryNodeConfig {
	return &InMemoryNodeConfig{
		TMConfig: tmcfg,
		ConsensusParams: abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxTxBytes:   1_000_000,   // 1MB,
				MaxDataBytes: 2_000_000,   // 2MB,
				MaxGas:       10_0000_000, // 10M gas
				TimeIotaMS:   100,         // 100ms
			},
		},
		GenesisMaxVMCycles: 10_000_000,
	}
}

// NewInMemoryNode creates an in-memory gnoland node. In this mode, the node does not
// persist any data and uses an in-memory database. The `InMemoryNodeConfig.TMConfig.RootDir`
// should point to the correct gno repository to load the stdlibs.
func NewInMemoryNode(logger log.Logger, cfg *InMemoryNodeConfig) (*node.Node, error) {
	if cfg.TMConfig == nil {
		return nil, fmt.Errorf("no `TMConfig` given")
	}

	if cfg.TMConfig.RootDir == "" {
		return nil, fmt.Errorf("`TMConfig.RootDir` is required but not provided")
	}

	// Create Identity
	nodekey := &p2p.NodeKey{PrivKey: ed25519.GenPrivKey()}
	pv := bft.NewMockPVWithParams(ed25519.GenPrivKey(), false, false)

	// Set up genesis with default values and additional validators
	gen := &bft.GenesisDoc{
		GenesisTime:     time.Now(),
		ChainID:         cfg.TMConfig.ChainID(),
		ConsensusParams: cfg.ConsensusParams,
		Validators: []bft.GenesisValidator{
			{
				Address: pv.GetPubKey().Address(),
				PubKey:  pv.GetPubKey(),
				Power:   10,
				Name:    "self",
			},
		},
	}
	gen.Validators = append(gen.Validators, cfg.GenesisValidator...)

	// XXX: Maybe let the user do this manually and pass it to genesisTXs
	txs, err := LoadPackages(cfg.Packages)
	if err != nil {
		return nil, fmt.Errorf("error loading genesis packages: %w", err)
	}

	// Combine loaded packages with provided genesis transactions
	txs = append(txs, cfg.GenesisTXs...)
	gen.AppState = GnoGenesisState{
		Balances: cfg.Balances,
		Txs:      txs,
	}

	// Initialize the application with the provided options
	gnoApp, err := NewAppWithOptions(&AppOptions{
		Logger:                logger,
		GnoRootDir:            cfg.TMConfig.RootDir,
		SkipFailingGenesisTxs: cfg.SkipFailingGenesisTxs,
		MaxCycles:             cfg.GenesisMaxVMCycles,
		DB:                    db.NewMemDB(),
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing new app: %w", err)
	}

	cfg.TMConfig.LocalApp = gnoApp

	// Setup app client creator
	appClientCreator := proxy.DefaultClientCreator(
		cfg.TMConfig.LocalApp,
		cfg.TMConfig.ProxyApp,
		cfg.TMConfig.ABCI,
		cfg.TMConfig.DBDir(),
	)

	// Create genesis factory
	genProvider := func() (*bft.GenesisDoc, error) {
		return gen, nil
	}

	// Create and return the in-memory node instance
	return node.NewNode(cfg.TMConfig,
		pv, nodekey,
		appClientCreator,
		genProvider,
		node.DefaultDBProvider,
		logger
	)
}

// LoadPackages loads and returns transactions from provided package paths.
func LoadPackages(pkgs []PackagePath) ([]std.Tx, error) {
	var txs []std.Tx
	for _, pkg := range pkgs {
		tx, err := pkg.Load()
		if err != nil {
			return nil, fmt.Errorf("error loading package from path %s: %w", pkg.Path, err)
		}
		txs = append(txs, tx...)
	}
	return txs, nil
}
