package integration

import (
	"errors"
	"fmt"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type NodeConfig struct {
	BFTConfig             *config.Config
	ConsensusParams       abci.ConsensusParams
	GenesisValidator      []bft.GenesisValidator
	Packages              []PackagePath
	Balances              []gnoland.Balance
	GenesisTXs            []std.Tx
	SkipFailingGenesisTxs bool
	GenesisMaxVMCycles    int64
}

func NewNode(logger log.Logger, cfg NodeConfig) (*node.Node, error) {
	// Setup setup testing config
	{
		if cfg.BFTConfig == nil {
			cfg.BFTConfig = config.TestConfig()
			cfg.BFTConfig.RPC.ListenAddress = "tcp://127.0.0.1:0"
			cfg.BFTConfig.P2P.ListenAddress = "tcp://127.0.0.1:0"
		}

		// XXX: we need to get ride of this, for now needed because of stdlib
		if cfg.BFTConfig.RootDir == "" {
			gnoRootDir := gnoland.MustGuessGnoRootDir()
			cfg.BFTConfig.SetRootDir(gnoRootDir)
		}
	}

	// generate node identity
	nodekey := &p2p.NodeKey{PrivKey: ed25519.GenPrivKey()}
	priv := bft.NewMockPVWithParams(nodekey.PrivKey, false, false)

	// Setup geeneis
	gen := &bft.GenesisDoc{}
	{

		gen.GenesisTime = time.Now()

		// cfg.chainID = "tendermint_test"
		gen.ChainID = cfg.BFTConfig.ChainID()

		// XXX(gfanton): Do we need some default here ?
		// if icfg.ConsensusParams.Block == nil {
		// 	icfg.ConsensusParams.Block = &abci.BlockParams{
		// 		MaxTxBytes:   1000000,  // 1MB,
		// 		MaxDataBytes: 2000000,  // 2MB,
		// 		MaxGas:       10000000, // 10M gas
		// 		TimeIotaMS:   100,      // 100ms
		// 	}
		// }
		gen.ConsensusParams = cfg.ConsensusParams

		pk := priv.GetPubKey()

		// start with self validator
		gen.Validators = []bft.GenesisValidator{
			{
				Address: pk.Address(),
				PubKey:  pk,
				Power:   10,
				Name:    "rootValidator",
			},
		}

		for _, validator := range cfg.GenesisValidator {
			gen.Validators = append(gen.Validators, validator)
		}
	}

	// XXX: maybe let the user do this manually and pass it to genesisTXs
	txs, err := LoadPackages(cfg.Packages)
	if err != nil {
		return nil, fmt.Errorf("uanble to load genesis packages: %w", err)
	}

	txs = append(txs, cfg.GenesisTXs...)

	gen.AppState = gnoland.GnoGenesisState{
		Balances: cfg.Balances,
		Txs:      txs,
	}

	gnoApp, err := gnoland.NewAppWithOptions(&gnoland.AppOptions{
		Logger:                logger,
		GnoRootDir:            cfg.BFTConfig.RootDir,
		SkipFailingGenesisTxs: cfg.SkipFailingGenesisTxs,
		MaxCycles:             cfg.GenesisMaxVMCycles,
		DB:                    db.NewMemDB(),
	})
	if err != nil {
		return nil, fmt.Errorf("error in creating new app: %w", err)
	}

	cfg.BFTConfig.LocalApp = gnoApp

	// Get app client creator.
	appClientCreator := proxy.DefaultClientCreator(
		cfg.BFTConfig.LocalApp,
		cfg.BFTConfig.ProxyApp,
		cfg.BFTConfig.ABCI,
		cfg.BFTConfig.DBDir(),
	)

	// Create genesis factory.
	genProvider := func() (*types.GenesisDoc, error) {
		return gen, nil
	}

	return node.NewNode(cfg.BFTConfig,
		priv, nodekey,
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

type PackagePath struct {
	Creator bft.Address
	Deposit std.Coins
	Fee     std.Fee
	Path    string
}

func (p PackagePath) Load() ([]std.Tx, error) {
	if p.Creator.IsZero() {
		return nil, errors.New("empty creator address")
	}

	if p.Path == "" {
		return nil, errors.New("empty package path")
	}

	// list all packages from target path
	pkgs, err := gnomod.ListPkgs(p.Path)
	if err != nil {
		return nil, fmt.Errorf("listing gno packages: %w", err)
	}

	// Sort packages by dependencies.
	sortedPkgs, err := pkgs.Sort()
	if err != nil {
		return nil, fmt.Errorf("sorting packages: %w", err)
	}

	// Filter out draft packages.
	nonDraftPkgs := sortedPkgs.GetNonDraftPkgs()
	txs := []std.Tx{}
	for _, pkg := range nonDraftPkgs {
		// Open files in directory as MemPackage.
		memPkg := gno.ReadMemPackage(pkg.Dir, pkg.Name)

		// Create transaction
		tx := std.Tx{
			Fee: p.Fee,
			Msgs: []std.Msg{
				vmm.MsgAddPackage{
					Creator: p.Creator,
					Package: memPkg,
					Deposit: p.Deposit,
				},
			},
		}

		tx.Signatures = make([]std.Signature, len(tx.GetSigners()))
		txs = append(txs, tx)
	}

	return txs, nil
}
