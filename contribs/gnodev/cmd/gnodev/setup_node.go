package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"

	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// setupDevNode initializes and returns a new DevNode.
func setupDevNode(
	ctx context.Context,
	logger *slog.Logger,
	cfg *devCfg,
	remitter emitter.Emitter,
	balances gnoland.Balances,
	pkgspath []gnodev.PackagePath,
) (*gnodev.Node, error) {
	// Load transactions.
	txs, err := parseTxs(cfg.txsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to load transactions: %w", err)
	}

	config := setupDevNodeConfig(cfg, balances, pkgspath, txs)
	return gnodev.NewDevNode(ctx, logger, remitter, config)
}

// setupDevNodeConfig creates and returns a new dev.NodeConfig.
func setupDevNodeConfig(
	cfg *devCfg,
	balances gnoland.Balances,
	pkgspath []gnodev.PackagePath,
	txs []std.Tx,
) *gnodev.NodeConfig {
	config := gnodev.DefaultNodeConfig(cfg.root)
	config.BalancesList = balances.List()
	config.PackagesPathList = pkgspath
	config.TMConfig.RPC.ListenAddress = resolveUnixOrTCPAddr(cfg.nodeRPCListenerAddr)
	config.NoReplay = cfg.noReplay
	config.MaxGasPerBlock = cfg.maxGas
	config.ChainID = cfg.chainId
	config.Txs = txs

	// other listeners
	config.TMConfig.P2P.ListenAddress = defaultDevOptions.nodeP2PListenerAddr
	config.TMConfig.ProxyApp = defaultDevOptions.nodeProxyAppListenerAddr

	return config
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
