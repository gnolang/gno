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

// NewInMemoryNode create an inMemeory gnoland node. In this mode, the node will
// not persist any data using an InMemory Database. For now the only indirect
// requirement is that `InMemoryNodeConfig.TMConfig.RootDir` is pointing to
// correct gno repository so that the node can load stdlibs
func NewInMemoryNode(logger log.Logger, cfg *InMemoryNodeConfig) (*node.Node, error) {
	if cfg.TMConfig == nil {
		return nil, fmt.Errorf("no TMConfig given")
	}

	// Create Identity
	nodekey := &p2p.NodeKey{PrivKey: ed25519.GenPrivKey()}
	pv := bft.NewMockPVWithParams(ed25519.GenPrivKey(), false, false)

	// Setup geeneis
	gen := &bft.GenesisDoc{}
	{
		gen.GenesisTime = time.Now()

		gen.ChainID = cfg.TMConfig.ChainID()

		gen.ConsensusParams = cfg.ConsensusParams

		// Register self first
		pk := pv.GetPubKey()
		gen.Validators = []bft.GenesisValidator{
			{
				Address: pk.Address(),
				PubKey:  pk,
				Power:   10,
				Name:    "self",
			},
		}

		for _, validator := range cfg.GenesisValidator {
			gen.Validators = append(gen.Validators, validator)
		}
	}

	// XXX: Maybe let the user do this manually and pass it to genesisTXs
	txs, err := LoadPackages(cfg.Packages)
	if err != nil {
		return nil, fmt.Errorf("uanble to load genesis packages: %w", err)
	}

	txs = append(txs, cfg.GenesisTXs...)

	gen.AppState = GnoGenesisState{
		Balances: cfg.Balances,
		Txs:      txs,
	}

	gnoApp, err := NewAppWithOptions(&AppOptions{
		Logger:                logger,
		GnoRootDir:            cfg.TMConfig.RootDir,
		SkipFailingGenesisTxs: cfg.SkipFailingGenesisTxs,
		MaxCycles:             cfg.GenesisMaxVMCycles,
		DB:                    db.NewMemDB(),
	})
	if err != nil {
		return nil, fmt.Errorf("error in creating new app: %w", err)
	}

	cfg.TMConfig.LocalApp = gnoApp

	// Get app client creator
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

	return node.NewNode(cfg.TMConfig,
		pv, nodekey,
		appClientCreator,
		genProvider,
		node.DefaultDBProvider,
		logger,
	)
}

func LoadPackages(pkgs []PackagePath) ([]std.Tx, error) {
	txs := []std.Tx{}
	for _, pkg := range pkgs {
		tx, err := pkg.Load()
		if err != nil {
			return nil, fmt.Errorf("unable to load packages: %w", err)
		}
		txs = append(txs, tx...)
	}
	return txs, nil
}
