package integration

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	tmcfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
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
func TestingInMemoryNode(t *testing.T, logger log.Logger, config *gnoland.InMemoryNodeConfig) (*node.Node, string) {
	t.Helper()

	node, err := gnoland.NewInMemoryNode(logger, config)
	require.NoError(t, err)

	err = node.Start()
	require.NoError(t, err)

	select {
	case <-waitForNodeReadiness(node):
	case <-time.After(time.Second * 6):
		require.FailNow(t, "timeout while waiting for the node to start")
	}

	return node, node.Config().RPC.ListenAddress
}

// DefaultTestingNodeConfig constructs the default in-memory node configuration for testing.
func DefaultTestingNodeConfig(t *testing.T, gnoroot string) *gnoland.InMemoryNodeConfig {
	t.Helper()

	tmconfig := DefaultTestingTMConfig(t, gnoroot)

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

func DefaultTestingGenesisConfig(t *testing.T, gnoroot string, self crypto.PubKey, tmconfig *tmcfg.Config) *bft.GenesisDoc {
	pkgCreator := crypto.MustAddressFromString(DefaultAccount_Address) // test1

	// Load genesis packages
	genesisPackagesTxs := LoadDefaultPackages(t, pkgCreator, gnoroot)

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
					Address: pkgCreator,
					Value:   std.MustParseCoins("10000000000000ugnot"),
				},
			},
			Txs: genesisPackagesTxs,
		},
	}
}

// LoadDefaultPackages loads the default packages for testing using a given creator address and gnoroot directory.
func LoadDefaultPackages(t *testing.T, creator bft.Address, gnoroot string) []std.Tx {
	t.Helper()

	exampleDir := filepath.Join(gnoroot, "examples")

	txs, err := gnoland.LoadPackages(gnoland.PackagePath{
		Creator: creator,
		Fee:     std.NewFee(50000, std.MustParseCoin("1000000ugnot")),
		Path:    exampleDir,
	})
	require.NoError(t, err)

	return txs
}

// LoadDefaultGenesisBalanceFile loads the default genesis balance file for testing.
func LoadDefaultGenesisBalanceFile(t *testing.T, gnoroot string) []gnoland.Balance {
	t.Helper()

	balanceFile := filepath.Join(gnoroot, "gno.land", "genesis", "genesis_balances.txt")

	genesisBalances, err := gnoland.LoadGenesisBalancesFile(balanceFile)
	require.NoError(t, err)

	return genesisBalances
}

// LoadDefaultGenesisTXsFile loads the default genesis transactions file for testing.
func LoadDefaultGenesisTXsFile(t *testing.T, chainid string, gnoroot string) []std.Tx {
	t.Helper()

	txsFile := filepath.Join(gnoroot, "gno.land", "genesis", "genesis_txs.txt")

	// NOTE: We dont care about giving a correct address here, as it's only for display
	// XXX: Do we care loading this TXs for testing ?
	genesisTXs, err := gnoland.LoadGenesisTxsFile(txsFile, chainid, "https://127.0.0.1:26657")
	require.NoError(t, err)

	return genesisTXs
}

// DefaultTestingTMConfig constructs the default Tendermint configuration for testing.
func DefaultTestingTMConfig(t *testing.T, gnoroot string) *tmcfg.Config {
	t.Helper()

	const defaultListner = "tcp://127.0.0.1:0"

	tmconfig := tmcfg.TestConfig().SetRootDir(gnoroot)
	tmconfig.Consensus.CreateEmptyBlocks = true
	tmconfig.Consensus.CreateEmptyBlocksInterval = time.Duration(0)
	tmconfig.RPC.ListenAddress = defaultListner
	tmconfig.P2P.ListenAddress = defaultListner
	return tmconfig
}

// waitForNodeReadiness waits until the node is ready, signaling via the EventNewBlock event.
// XXX: This should be replace by https://github.com/gnolang/gno/pull/1216
func waitForNodeReadiness(n *node.Node) <-chan struct{} {
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
