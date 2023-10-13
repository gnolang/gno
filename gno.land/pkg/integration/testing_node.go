package integration

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
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
func TestingInMemoryNode(t *testing.T, logger log.Logger, config *TestingNodeConfig) (*node.Node, string) {
	t.Helper()

	node, err := NewTestingNode(logger, config)
	require.NoError(t, err)

	err = node.Start()
	require.NoError(t, err)

	// XXX: This should be replace by https://github.com/gnolang/gno/pull/1216
	//---
	// Wait for first block by waiting for `EventNewBlock` event.
	const listenerID = "testing_listener"

	nb := make(chan struct{}, 1)
	node.EventSwitch().AddListener(listenerID, func(ev events.Event) {
		if _, ok := ev.(bft.EventNewBlock); ok {
			select {
			case nb <- struct{}{}:
			default:
			}
		}
	})

	if node.BlockStore().Height() == 0 {
		select {
		case <-nb: // ok
		case <-time.After(time.Second * 6):
			t.Fatal("timeout while waiting for the node to start")
		}
	}

	node.EventSwitch().RemoveListener(listenerID)
	// ---

	return node, node.Config().RPC.ListenAddress
}

func DefaultTestingNodeConfig(t *testing.T, gnoroot string) *TestingNodeConfig {
	t.Helper()

	bftconfig := DefaultTestingBFTConfig(t, gnoroot)
	return &TestingNodeConfig{
		Balances:        LoadDefaultGenesisBalanceFile(t, gnoroot),
		GenesisTXs:      LoadDefaultGenesisTXsFile(t, bftconfig.ChainID(), gnoroot),
		ConsensusParams: DefaultConsensusParams(t),
		BFTConfig:       bftconfig,
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

	// NOTE: we dont care about giving a correct address here, as it's only visual
	// XXX: do we care loading this file ?
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

func DefaultTestingBFTConfig(t *testing.T, gnoroot string) *config.Config {
	t.Helper()

	const defaultListner = "tcp://127.0.0.1:0"

	bftconfig := config.TestConfig().SetRootDir(gnoroot)
	bftconfig.Consensus.CreateEmptyBlocks = true
	bftconfig.Consensus.CreateEmptyBlocksInterval = time.Duration(0)
	bftconfig.RPC.ListenAddress = defaultListner
	bftconfig.P2P.ListenAddress = defaultListner
	return bftconfig
}
