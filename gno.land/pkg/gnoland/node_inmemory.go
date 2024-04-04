package gnoland

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	tmcfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type InMemoryNodeConfig struct {
	PrivValidator         bft.PrivValidator // identity of the validator
	Genesis               *bft.GenesisDoc
	TMConfig              *tmcfg.Config
	SkipFailingGenesisTxs bool
}

// NewMockedPrivValidator generate a new key
func NewMockedPrivValidator() bft.PrivValidator {
	return bft.NewMockPVWithParams(ed25519.GenPrivKey(), false, false)
}

// NewInMemoryNodeConfig creates a default configuration for an in-memory node.
func NewDefaultGenesisConfig(pk crypto.PubKey, chainid string) *bft.GenesisDoc {
	return &bft.GenesisDoc{
		GenesisTime: time.Now(),
		ChainID:     chainid,
		ConsensusParams: abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxTxBytes:   1_000_000,   // 1MB,
				MaxDataBytes: 2_000_000,   // 2MB,
				MaxGas:       100_000_000, // 10M gas
				TimeIotaMS:   100,         // 100ms
			},
		},
		AppState: &GnoGenesisState{
			Balances: []Balance{},
			Txs:      []std.Tx{},
		},
	}
}

func NewDefaultTMConfig(rootdir string) *tmcfg.Config {
	// We use `TestConfig` here otherwise ChainID will be empty, and
	// there is no other way to update it than using a config file
	return tmcfg.TestConfig().SetRootDir(rootdir)
}

// NewInMemoryNodeConfig creates a default configuration for an in-memory node.
func NewDefaultInMemoryNodeConfig(rootdir string) *InMemoryNodeConfig {
	tm := NewDefaultTMConfig(rootdir)

	// Create Mocked Identity
	pv := NewMockedPrivValidator()
	genesis := NewDefaultGenesisConfig(pv.GetPubKey(), tm.ChainID())

	// Add self as validator
	self := pv.GetPubKey()
	genesis.Validators = []bft.GenesisValidator{
		{
			Address: self.Address(),
			PubKey:  self,
			Power:   10,
			Name:    "self",
		},
	}

	return &InMemoryNodeConfig{
		PrivValidator: pv,
		TMConfig:      tm,
		Genesis:       genesis,
	}
}

func (cfg *InMemoryNodeConfig) validate() error {
	if cfg.PrivValidator == nil {
		return fmt.Errorf("`PrivValidator` is required but not provided")
	}

	if cfg.TMConfig == nil {
		return fmt.Errorf("`TMConfig` is required but not provided")
	}

	if cfg.TMConfig.RootDir == "" {
		return fmt.Errorf("`TMConfig.RootDir` is required to locate `stdlibs` directory")
	}

	return nil
}

// NewInMemoryNode creates an in-memory gnoland node. In this mode, the node does not
// persist any data and uses an in-memory database. The `InMemoryNodeConfig.TMConfig.RootDir`
// should point to the correct gno repository to load the stdlibs.
func NewInMemoryNode(logger *slog.Logger, cfg *InMemoryNodeConfig) (*node.Node, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config error: %w", err)
	}

	// Initialize the application with the provided options
	gnoApp, err := NewAppWithOptions(&AppOptions{
		Logger:                logger,
		GnoRootDir:            cfg.TMConfig.RootDir,
		SkipFailingGenesisTxs: cfg.SkipFailingGenesisTxs,
		DB:                    memdb.NewMemDB(),
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
	genProvider := func() (*bft.GenesisDoc, error) { return cfg.Genesis, nil }

	dbProvider := func(*node.DBContext) (db.DB, error) { return memdb.NewMemDB(), nil }

	// generate p2p node identity
	// XXX: do we need to configur
	nodekey := &p2p.NodeKey{PrivKey: ed25519.GenPrivKey()}

	// Create and return the in-memory node instance
	return node.NewNode(cfg.TMConfig,
		cfg.PrivValidator, nodekey,
		appClientCreator,
		genProvider,
		dbProvider,
		logger,
	)
}

// GetNodeReadiness waits until the node is ready, signaling via the EventNewBlock event.
// XXX: This should be replace by https://github.com/gnolang/gno/pull/1216
func GetNodeReadiness(n *node.Node) <-chan struct{} {
	const listenerID = "first_block_listener"

	var once sync.Once

	nb := make(chan struct{})
	ready := func() {
		close(nb)
		n.EventSwitch().RemoveListener(listenerID)
	}

	n.EventSwitch().AddListener(listenerID, func(ev events.Event) {
		if _, ok := ev.(bft.EventNewBlock); ok {
			once.Do(ready)
		}
	})

	if n.BlockStore().Height() > 0 {
		once.Do(ready)
	}

	return nb
}
