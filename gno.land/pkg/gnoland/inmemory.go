package gnoland

import (
	"errors"
	"fmt"
	"time"

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

type InMemoryConfig struct {
	RootDir               string
	ConsensusParams       abci.ConsensusParams
	GenesisValidator      []bft.GenesisValidator
	Packages              []PackagePath
	Balances              []Balance
	GenesisTXs            []std.Tx
	SkipFailingGenesisTxs bool
	GenesisMaxVMCycles    int64
}

func (im *InMemoryConfig) loadPackages() ([]std.Tx, error) {
	txs := []std.Tx{}
	for _, pkg := range im.Packages {
		tx, err := pkg.load()
		if err != nil {
			return nil, fmt.Errorf("unable to load packages: %w", err)
		}
		txs = append(txs, tx...)

	}
	return txs, nil
}

func NewInMemory(logger log.Logger, icfg InMemoryConfig) (*node.Node, error) {
	if icfg.RootDir == "" {
		// XXX: Should return an error here ?
		icfg.RootDir = MustGuessGnoRootDir()
	}

	// Setup testing config
	cfg := config.TestConfig().SetRootDir(icfg.RootDir)
	{
		cfg.EnsureDirs()
		cfg.Consensus.CreateEmptyBlocks = true
		cfg.Consensus.CreateEmptyBlocksInterval = time.Duration(0)
		cfg.RPC.ListenAddress = "tcp://127.0.0.1:0"
		cfg.P2P.ListenAddress = "tcp://127.0.0.1:0"
	}

	// use mocked pv
	nodekey := &p2p.NodeKey{
		PrivKey: ed25519.GenPrivKey(),
	}
	priv := bft.NewMockPVWithParams(nodekey.PrivKey, false, false)

	// setup geeneis
	gen := &bft.GenesisDoc{}
	{

		gen.GenesisTime = time.Now()

		// cfg.chainID = "tendermint_test"
		gen.ChainID = cfg.ChainID()

		// XXX(gfanton): Is some a default needed here ?
		// if icfg.ConsensusParams.Block == nil {
		// 	icfg.ConsensusParams.Block = &abci.BlockParams{
		// 		// TODO: update limits based on config
		// 		MaxTxBytes:   1000000,  // 1MB,
		// 		MaxDataBytes: 2000000,  // 2MB,
		// 		MaxGas:       10000000, // 10M gas
		// 		TimeIotaMS:   100,      // 100ms
		// 	}
		// }
		gen.ConsensusParams = icfg.ConsensusParams

		pk := priv.GetPubKey()
		gen.Validators = []bft.GenesisValidator{
			{
				Address: pk.Address(),
				PubKey:  pk,
				Power:   10,
				Name:    "testvalidator",
			},
		}

		for _, validator := range icfg.GenesisValidator {
			gen.Validators = append(gen.Validators, validator)
		}
	}

	txs, err := icfg.loadPackages()
	if err != nil {
		return nil, fmt.Errorf("uanble to load genesis packages: %w", err)
	}

	txs = append(txs, icfg.GenesisTXs...)

	gen.AppState = GnoGenesisState{
		Balances: Balances(icfg.Balances).Strings(),
		Txs:      txs,
	}

	gnoApp, err := NewAppWithOptions(&AppOptions{
		Logger:                logger,
		GnoRootDir:            icfg.RootDir,
		SkipFailingGenesisTxs: icfg.SkipFailingGenesisTxs,
		MaxCycles:             icfg.GenesisMaxVMCycles,
		DB:                    db.NewMemDB(),
	})
	if err != nil {
		return nil, fmt.Errorf("error in creating new app: %w", err)
	}

	cfg.LocalApp = gnoApp

	// Get app client creator.
	appClientCreator := proxy.DefaultClientCreator(
		cfg.LocalApp,
		cfg.ProxyApp,
		cfg.ABCI,
		cfg.DBDir(),
	)

	// Create genesis factory.
	genProvider := func() (*types.GenesisDoc, error) {
		return gen, nil
	}

	return node.NewNode(cfg,
		priv, nodekey,
		appClientCreator,
		genProvider,
		node.DefaultDBProvider,
		logger,
	)
}

type PackagePath struct {
	Creator bft.Address
	Deposit std.Coins
	Fee     std.Fee
	Path    string
}

func (p PackagePath) load() ([]std.Tx, error) {
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
		// open files in directory as MemPackage.
		memPkg := gno.ReadMemPackage(pkg.Dir, pkg.Name)

		// create transaction
		tx := std.Tx{
			Fee: p.Fee,
			Msgs: []std.Msg{
				vmm.MsgAddPackage{
					Creator: p.Creator,
					Package: memPkg,
					// XXX: add deposit option
					Deposit: p.Deposit,
				},
			},
		}

		tx.Signatures = make([]std.Signature, len(tx.GetSigners()))
		txs = append(txs, tx)
	}

	return txs, nil
}

func WaitForReadiness(n *node.Node) <-chan struct{} {
	go func() {

	}()
}

// func loadGenesisTxs(
// 	path string,
// 	chainID string,
// 	genesisRemote string,
// ) []std.Tx {
// 	txs := []std.Tx{}
// 	txsBz := osm.MustReadFile(path)
// 	txsLines := strings.Split(string(txsBz), "\n")
// 	for _, txLine := range txsLines {
// 		if txLine == "" {
// 			continue // skip empty line
// 		}

// 		// patch the TX
// 		txLine = strings.ReplaceAll(txLine, "%%CHAINID%%", chainID)
// 		txLine = strings.ReplaceAll(txLine, "%%REMOTE%%", genesisRemote)

// 		var tx std.Tx
// 		amino.MustUnmarshalJSON([]byte(txLine), &tx)
// 		txs = append(txs, tx)
// 	}

// 	return txs
// }

// func setupTestingGenesis(gnoDataDir string, cfg *config.Config, icfg *IntegrationConfig) error {
// 	genesisFilePath := filepath.Join(gnoDataDir, cfg.Genesis)
// 	osm.EnsureDir(filepath.Dir(genesisFilePath), 0o700)
// 	if !osm.FileExists(genesisFilePath) {
// 		genesisTxs := loadGenesisTxs(icfg.GenesisTxsFile, icfg.ChainID, icfg.GenesisRemote)
// 		pvPub := priv.GetPubKey()

// 		gen := &bft.GenesisDoc{
// 			GenesisTime: time.Now(),
// 			ChainID:     icfg.ChainID,
// 			ConsensusParams: abci.ConsensusParams{
// 				Block: &abci.BlockParams{
// 					// TODO: update limits.
// 					MaxTxBytes:   1000000,  // 1MB,
// 					MaxDataBytes: 2000000,  // 2MB,
// 					MaxGas:       10000000, // 10M gas
// 					TimeIotaMS:   100,      // 100ms
// 				},
// 			},
// 			Validators: []bft.GenesisValidator{
// 				{
// 					Address: pvPub.Address(),
// 					PubKey:  pvPub,
// 					Power:   10,
// 					Name:    "testvalidator",
// 				},
// 			},
// 		}

// 		// Load distribution.
// 		balances := loadGenesisBalances(icfg.GenesisBalancesFile)

// 		// Load initial packages from examples.
// 		// XXX: we should be able to config this
// 		test1 := crypto.MustAddressFromString(test1Addr)
// 		txs := []std.Tx{}

// 		// List initial packages to load from examples.
// 		// println(filepath.Join(gnoRootDir, "examples"))

// 		// load genesis txs from file.
// 		txs = append(txs, genesisTxs...)

// 		// construct genesis AppState.
// 		gen.AppState = GnoGenesisState{
// 			Balances: balances,
// 			Txs:      txs,
// 		}

// 		writeGenesisFile(gen, genesisFilePath)
// 	}

// 	return nil
// }

// func loadGenesisBalances(path string) []string {
// 	// each balance is in the form: g1xxxxxxxxxxxxxxxx=100000ugnot
// 	balances := []string{}
// 	content := osm.MustReadFile(path)
// 	lines := strings.Split(string(content), "\n")
// 	for _, line := range lines {
// 		line = strings.TrimSpace(line)

// 		// remove comments.
// 		line = strings.Split(line, "#")[0]
// 		line = strings.TrimSpace(line)

// 		// skip empty lines.
// 		if line == "" {
// 			continue
// 		}

// 		parts := strings.Split(line, "=")
// 		if len(parts) != 2 {
// 			panic("invalid genesis_balance line: " + line)
// 		}

// 		balances = append(balances, line)
// 	}
// 	return balances
// }

// func writeGenesisFile(gen *bft.GenesisDoc, filePath string) {
// 	err := gen.SaveAs(filePath)
// 	if err != nil {
// 		panic(err)
// 	}
// }
