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
	DefaultAccountName    = "test1"
	DefaultAccountAddress = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	DefaultAccountSeed    = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
)

// Should return an already starting node
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

func DefaultTestingNodeConfig(t *testing.T, gnoroot string) *gnoland.InMemoryNodeConfig {
	t.Helper()

	tmconfig := DefaultTestingTMConfig(t, gnoroot)
	return &gnoland.InMemoryNodeConfig{
		Balances:        LoadDefaultGenesisBalanceFile(t, gnoroot),
		GenesisTXs:      LoadDefaultGenesisTXsFile(t, tmconfig.ChainID(), gnoroot),
		ConsensusParams: DefaultConsensusParams(t),
		TMConfig:        tmconfig,
		Packages:        LoadDefaultPackages(t, crypto.MustAddressFromString(DefaultAccountAddress), gnoroot),
	}
}

func LoadDefaultPackages(t *testing.T, creator bft.Address, gnoroot string) []gnoland.PackagePath {
	t.Helper()

	exampleDir := filepath.Join(gnoroot, "examples")

	pkgs := gnoland.PackagePath{
		Creator: creator,
		Fee:     std.NewFee(50000, std.MustParseCoin("1000000ugnot")),
		Path:    exampleDir,
	}

	return []gnoland.PackagePath{pkgs}
}

func LoadDefaultGenesisBalanceFile(t *testing.T, gnoroot string) []gnoland.Balance {
	t.Helper()

	balanceFile := filepath.Join(gnoroot, "gno.land", "genesis", "genesis_balances.txt")

	genesisBalances, err := gnoland.LoadGenesisBalancesFile(balanceFile)
	require.NoError(t, err)

	return genesisBalances
}

func LoadDefaultGenesisTXsFile(t *testing.T, chainid string, gnoroot string) []std.Tx {
	t.Helper()

	txsFile := filepath.Join(gnoroot, "gno.land", "genesis", "genesis_txs.txt")

	// NOTE: We dont care about giving a correct address here, as it's only for display
	// XXX: Do we care loading this TXs for testing ?
	genesisTXs, err := gnoland.LoadGenesisTxsFile(txsFile, chainid, "https://127.0.0.1:26657")
	require.NoError(t, err)

	return genesisTXs
}

func DefaultConsensusParams(t *testing.T) abci.ConsensusParams {
	t.Helper()

	return abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxTxBytes:   1_000_000,  // 1MB,
			MaxDataBytes: 2_000_000,  // 2MB,
			MaxGas:       10_000_000, // 10M gas
			TimeIotaMS:   100,        // 100ms
		},
	}
}

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
