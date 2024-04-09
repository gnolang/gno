package integration

import (
	"log/slog"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	tmcfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/jaekwon/testify/require"
)

const (
	DefaultAccount_Name    = "test1"
	DefaultAccount_Address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	DefaultAccount_Seed    = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
)

// TestingInMemoryNode initializes and starts an in-memory node for testing.
// It returns the node instance and its RPC remote address.
func TestingInMemoryNode(t TestingTS, logger *slog.Logger, config *gnoland.InMemoryNodeConfig) (*node.Node, string) {
	node, err := gnoland.NewInMemoryNode(logger, config)
	require.NoError(t, err)

	err = node.Start()
	require.NoError(t, err)

	select {
	case <-node.Ready():
	case <-time.After(time.Second * 10):
		require.FailNow(t, "timeout while waiting for the node to start")
	}

	return node, node.Config().RPC.ListenAddress
}

// TestingNodeConfig constructs an in-memory node configuration
// with default packages and genesis transactions already loaded.
// It will return the default creator address of the loaded packages.
func TestingNodeConfig(t TestingTS, gnoroot string) (*gnoland.InMemoryNodeConfig, bft.Address) {
	cfg := TestingMinimalNodeConfig(t, gnoroot)

	creator := crypto.MustAddressFromString(DefaultAccount_Address) // test1

	balances := LoadDefaultGenesisBalanceFile(t, gnoroot)
	txs := []std.Tx{}
	txs = append(txs, LoadDefaultPackages(t, creator, gnoroot)...)
	txs = append(txs, LoadDefaultGenesisTXsFile(t, cfg.Genesis.ChainID, gnoroot)...)

	cfg.Genesis.AppState = gnoland.GnoGenesisState{
		Balances: balances,
		Txs:      txs,
	}

	return cfg, creator
}

// TestingMinimalNodeConfig constructs the default minimal in-memory node configuration for testing.
func TestingMinimalNodeConfig(t TestingTS, gnoroot string) *gnoland.InMemoryNodeConfig {
	tmconfig := DefaultTestingTMConfig(gnoroot)

	// Create Mocked Identity
	pv := gnoland.NewMockedPrivValidator()

	// Generate genesis config
	genesis := DefaultTestingGenesisConfig(t, gnoroot, pv.GetPubKey(), tmconfig)

	return &gnoland.InMemoryNodeConfig{
		PrivValidator: pv,
		Genesis:       genesis,
		TMConfig:      tmconfig,
	}
}

func DefaultTestingGenesisConfig(t TestingTS, gnoroot string, self crypto.PubKey, tmconfig *tmcfg.Config) *bft.GenesisDoc {
	return &bft.GenesisDoc{
		GenesisTime: time.Now(),
		ChainID:     tmconfig.ChainID(),
		ConsensusParams: abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxTxBytes:   1_000_000,   // 1MB,
				MaxDataBytes: 2_000_000,   // 2MB,
				MaxGas:       10_0000_000, // 10M gas
				TimeIotaMS:   100,         // 100ms
			},
		},
		Validators: []bft.GenesisValidator{
			{
				Address: self.Address(),
				PubKey:  self,
				Power:   10,
				Name:    "self",
			},
		},
		AppState: gnoland.GnoGenesisState{
			Balances: []gnoland.Balance{
				{
					Address: crypto.MustAddressFromString(DefaultAccount_Address),
					Amount:  std.MustParseCoins("10000000000000ugnot"),
				},
			},
			Txs: []std.Tx{},
		},
	}
}

// LoadDefaultPackages loads the default packages for testing using a given creator address and gnoroot directory.
func LoadDefaultPackages(t TestingTS, creator bft.Address, gnoroot string) []std.Tx {
	examplesDir := filepath.Join(gnoroot, "examples")

	defaultFee := std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
	txs, err := gnoland.LoadPackagesFromDir(examplesDir, creator, defaultFee, nil)
	require.NoError(t, err)

	return txs
}

// LoadDefaultGenesisBalanceFile loads the default genesis balance file for testing.
func LoadDefaultGenesisBalanceFile(t TestingTS, gnoroot string) []gnoland.Balance {
	balanceFile := filepath.Join(gnoroot, "gno.land", "genesis", "genesis_balances.txt")

	genesisBalances, err := gnoland.LoadGenesisBalancesFile(balanceFile)
	require.NoError(t, err)

	return genesisBalances
}

// LoadDefaultGenesisTXsFile loads the default genesis transactions file for testing.
func LoadDefaultGenesisTXsFile(t TestingTS, chainid string, gnoroot string) []std.Tx {
	txsFile := filepath.Join(gnoroot, "gno.land", "genesis", "genesis_txs.jsonl")

	// NOTE: We dont care about giving a correct address here, as it's only for display
	// XXX: Do we care loading this TXs for testing ?
	genesisTXs, err := gnoland.LoadGenesisTxsFile(txsFile, chainid, "https://127.0.0.1:26657")
	require.NoError(t, err)

	return genesisTXs
}

// DefaultTestingTMConfig constructs the default Tendermint configuration for testing.
func DefaultTestingTMConfig(gnoroot string) *tmcfg.Config {
	const defaultListner = "tcp://127.0.0.1:0"

	tmconfig := tmcfg.TestConfig().SetRootDir(gnoroot)
	tmconfig.Consensus.WALDisabled = true
	tmconfig.Consensus.CreateEmptyBlocks = true
	tmconfig.Consensus.CreateEmptyBlocksInterval = time.Duration(0)
	tmconfig.RPC.ListenAddress = defaultListner
	tmconfig.P2P.ListenAddress = defaultListner
	return tmconfig
}
