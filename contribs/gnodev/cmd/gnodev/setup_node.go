package main

import (
	"fmt"
	"log/slog"
	"net"
	"strings"

	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	emitter "github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// setupDevNodeConfig creates and returns a new dev.NodeConfig.
func setupDevNodeConfig(
	cfg *devCfg,
	logger *slog.Logger,
	emitter emitter.Emitter,
	balances gnoland.Balances,
	pkgspath []gnodev.PackagePath,
) *gnodev.NodeConfig {
	config := gnodev.DefaultNodeConfig(cfg.root)
	if cfg.genesisFile != "" {
		var err error
		if config.InitialTxs, err = extractTxsFromGenesisFile(cfg.genesisFile); err != nil {
			logger.Error("unable to load genesis file", "path", cfg.genesisFile, "err", err)
		} else {
			logger.Info("genesis file loaded", "path", cfg.genesisFile, "txs", len(config.InitialTxs))
		}
	}

	config.Logger = logger
	config.Emitter = emitter
	config.BalancesList = balances.List()
	config.PackagesPathList = pkgspath
	config.TMConfig.RPC.ListenAddress = resolveUnixOrTCPAddr(cfg.nodeRPCListenerAddr)
	config.NoReplay = cfg.noReplay
	config.MaxGasPerBlock = cfg.maxGas
	config.ChainID = cfg.chainId

	// other listeners
	config.TMConfig.P2P.ListenAddress = defaultDevOptions.nodeP2PListenerAddr
	config.TMConfig.ProxyApp = defaultDevOptions.nodeProxyAppListenerAddr

	return config
}

func extractTxsFromGenesisFile(path string) (txs []std.Tx, err error) {
	doc, err := types.GenesisDocFromFile(path)
	if err != nil {
		return []std.Tx{}, fmt.Errorf("unable to parse doc file: %w", err)
	}

	state, ok := doc.AppState.(gnoland.GnoGenesisState)
	if !ok {
		return []std.Tx{}, fmt.Errorf("invalid `GnoGenesisState` app state")
	}

	return state.Txs, nil
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
