package integration

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/jaekwon/testify/require"
)

func DefaultTestingConfig(t *testing.T, gnoroot string) *NodeConfig {
	txsFile := filepath.Join(gnoroot, "gno.land", "genesis", "genesis_txs.txt")
	balanceFile := filepath.Join(gnoroot, "gno.land", "genesis", "genesis_balances.txt")
	exampleDir := filepath.Join(gnoroot, "examples")

	genesisBalances, err := gnoland.LoadGenesisBalancesFile(balanceFile)
	require.NoError(t, err)

	bftconfig := config.TestConfig().SetRootDir(gnoroot)
	bftconfig.Consensus.CreateEmptyBlocks = true
	bftconfig.Consensus.CreateEmptyBlocksInterval = time.Duration(0)
	bftconfig.RPC.ListenAddress = "tcp://127.0.0.1:0"
	bftconfig.P2P.ListenAddress = "tcp://127.0.0.1:0"

	// NOTE: we dont care about giving a correct address here, as it's only visual
	// XXX: do we care loading this file ?
	genesisTXs, err := gnoland.LoadGenesisTxsFile(txsFile, bftconfig.ChainID(), bftconfig.RPC.ListenAddress)
	require.NoError(t, err)

	// Load example packages
	pkgs := PackagePath{
		Creator: crypto.MustAddressFromString(test1Addr),
		Fee:     std.NewFee(50000, std.MustParseCoin("1000000ugnot")),
		Path:    exampleDir,
	}

	config := NodeConfig{
		Balances:   genesisBalances,
		GenesisTXs: genesisTXs,
		BFTConfig:  bftconfig,
		Packages:   []PackagePath{pkgs},
	}

	config.ConsensusParams.Block = &abci.BlockParams{
		MaxTxBytes:   1000000,  // 1MB,
		MaxDataBytes: 2000000,  // 2MB,
		MaxGas:       10000000, // 10M gas
		TimeIotaMS:   100,      // 100ms
	}

	return &config
}

// Should return an already starting node
func TestingInMemoryNode(t *testing.T, logger log.Logger, config *NodeConfig) *node.Node {
	node, err := NewNode(logger, *config)
	require.NoError(t, err)

	err = node.Start()
	require.NoError(t, err)

	// XXX: This should be replace by https://github.com/gnolang/gno/pull/1216
	const listenerID = "testing_listener"

	// Wait for first block by waiting for `EventNewBlock` event.
	nb := make(chan struct{}, 1)
	node.EventSwitch().AddListener(listenerID, func(ev events.Event) {
		if _, ok := ev.(types.EventNewBlock); ok {
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

	return node
}
