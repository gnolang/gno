package gnoland

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	tmcfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

type InMemoryNodeConfig struct {
	PrivValidator              bft.PrivValidator // identity of the validator
	Genesis                    *bft.GenesisDoc
	TMConfig                   *tmcfg.Config
	DB                         db.DB     // will be initialized if nil
	VMOutput                   io.Writer // optional
	SkipGenesisSigVerification bool

	// If StdlibDir not set, then it's filepath.Join(TMConfig.RootDir, "gnovm", "stdlibs")
	InitChainerConfig
}

// NewDefaultGenesisConfig creates a default configuration for an in-memory node.
func NewDefaultGenesisConfig(chainid, chaindomain string) *bft.GenesisDoc {
	return &bft.GenesisDoc{
		GenesisTime: time.Now(),
		ChainID:     chainid,
		ConsensusParams: abci.ConsensusParams{
			Block: defaultBlockParams(),
		},
		AppState: &GnoGenesisState{
			Balances: []Balance{},
			Txs:      []TxWithMetadata{},
			VM: vm.GenesisState{
				Params: vm.Params{
					ChainDomain: chaindomain,
				},
			},
		},
	}
}

func defaultBlockParams() *abci.BlockParams {
	return &abci.BlockParams{
		MaxTxBytes:   1_000_000,     // 1MB,
		MaxDataBytes: 2_000_000,     // 2MB,
		MaxGas:       3_000_000_000, // 3B gas
		TimeIotaMS:   100,           // 100ms
	}
}

func NewDefaultTMConfig(rootdir string) *tmcfg.Config {
	// We use `TestConfig` here otherwise ChainID will be empty, and
	// there is no other way to update it than using a config file
	return tmcfg.TestConfig().SetRootDir(rootdir)
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

	if cfg.GenesisTxResultHandler == nil {
		return fmt.Errorf("`GenesisTxHandler` is required but not provided")
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

	evsw := events.NewEventSwitch()

	if cfg.StdlibDir == "" {
		cfg.StdlibDir = filepath.Join(cfg.TMConfig.RootDir, "gnovm", "stdlibs")
	}
	// initialize db if nil
	if cfg.DB == nil {
		cfg.DB = memdb.NewMemDB()
	}

	// Initialize the application with the provided options
	gnoApp, err := NewAppWithOptions(&AppOptions{
		Logger:                     logger,
		DB:                         cfg.DB,
		EventSwitch:                evsw,
		InitChainerConfig:          cfg.InitChainerConfig,
		VMOutput:                   cfg.VMOutput,
		SkipGenesisSigVerification: cfg.SkipGenesisSigVerification,
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

	dbProvider := func(*node.DBContext) (db.DB, error) { return cfg.DB, nil }

	// Generate p2p node identity
	nodekey := &types.NodeKey{PrivKey: ed25519.GenPrivKey()}

	// Create and return the in-memory node instance
	return node.NewNode(
		cfg.TMConfig,
		cfg.PrivValidator,
		nodekey,
		appClientCreator,
		genProvider,
		dbProvider,
		evsw,
		logger,
	)
}
