package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"

	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

// setupDevNode initializes and returns a new DevNode.
func setupDevNode(ctx context.Context, devCfg *devCfg, nodeConfig *gnodev.NodeConfig, path string) (*gnodev.Node, error) {
	logger := nodeConfig.Logger

	if devCfg.txsFile != "" { // Load txs files
		var err error
		nodeConfig.InitialTxs, err = gnoland.ReadGenesisTxs(ctx, devCfg.txsFile)
		if err != nil {
			return nil, fmt.Errorf("unable to load transactions: %w", err)
		}
	} else if devCfg.genesisFile != "" { // Load genesis file
		state, err := extractAppStateFromGenesisFile(devCfg.genesisFile)
		if err != nil {
			return nil, fmt.Errorf("unable to load genesis file %q: %w", devCfg.genesisFile, err)
		}

		// Override balances and txs
		nodeConfig.BalancesList = state.Balances

		stateTxs := state.Txs
		nodeConfig.InitialTxs = make([]gnoland.TxWithMetadata, len(stateTxs))

		for index, nodeTx := range stateTxs {
			nodeConfig.InitialTxs[index] = nodeTx
		}

		logger.Info("genesis file loaded", "path", devCfg.genesisFile, "txs", len(stateTxs))
	}

	return gnodev.NewDevNode(ctx, nodeConfig, path)
}

// setupDevNodeConfig creates and returns a new dev.NodeConfig.
func setupDevNodeConfig(
	cfg *devCfg,
	logger *slog.Logger,
	emitter emitter.Emitter,
	balances gnoland.Balances,
	loader *packages.Loader,
) *gnodev.NodeConfig {
	config := gnodev.DefaultNodeConfig(cfg.root)
	config.Loader = loader

	config.Logger = logger
	config.Emitter = emitter
	config.BalancesList = balances.List()
	// config.PackagesModifier = pkgspath
	config.TMConfig.RPC.ListenAddress = resolveUnixOrTCPAddr(cfg.nodeRPCListenerAddr)
	config.NoReplay = cfg.noReplay
	config.MaxGasPerBlock = cfg.maxGas
	config.ChainID = cfg.chainId

	// other listeners
	config.TMConfig.P2P.ListenAddress = defaultDevOptions.nodeP2PListenerAddr
	config.TMConfig.ProxyApp = defaultDevOptions.nodeProxyAppListenerAddr

	return config
}

func extractAppStateFromGenesisFile(path string) (*gnoland.GnoGenesisState, error) {
	doc, err := types.GenesisDocFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to parse doc file: %w", err)
	}

	state, ok := doc.AppState.(gnoland.GnoGenesisState)
	if !ok {
		return nil, fmt.Errorf("invalid `GnoGenesisState` app state")
	}

	return &state, nil
}

func resolveUnixOrTCPAddr(in string) (out string) {
	var err error
	var addr net.Addr

	if strings.HasPrefix(in, "unix://") {
		in = strings.TrimPrefix(in, "unix://")
		if addr, err := net.ResolveUnixAddr("unix", in); err == nil {
			return fmt.Sprintf("%s://%s", addr.Network(), addr.String())
		}

		err = fmt.Errorf("unable to resolve unix address `unix://%s`: %w", in, err)
	} else { // don't bother to checking prefix
		in = strings.TrimPrefix(in, "tcp://")
		if addr, err = net.ResolveTCPAddr("tcp", in); err == nil {
			return fmt.Sprintf("%s://%s", addr.Network(), addr.String())
		}

		err = fmt.Errorf("unable to resolve tcp address `tcp://%s`: %w", in, err)
	}

	panic(err)
}
