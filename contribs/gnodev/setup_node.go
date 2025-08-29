package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"slices"
	"strings"

	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// extractDependenciesFromTxs extracts dependencies from transactions and adds them to the paths slice and config.BalancesList.
func extractDependenciesFromTxs(nodeConfig *gnodev.NodeConfig, paths *[]string) {
	var defaultPremineBalance = std.Coins{std.NewCoin(ugnot.Denom, 10e12)}

	for _, tx := range nodeConfig.InitialTxs {
		for _, msg := range tx.Tx.Msgs {
			// TODO: Support MsgRun
			callMsg, ok := msg.(vm.MsgCall)
			if !ok {
				continue
			}
			// Add package path to paths slice if not already present
			if !slices.Contains(*paths, callMsg.PkgPath) {
				*paths = append(*paths, callMsg.PkgPath)
			}

			// Check if address exists in config.BalancesList
			addressExists := false
			for _, balance := range nodeConfig.BalancesList {
				if balance.Address == callMsg.Caller {
					addressExists = true
					break
				}
			}

			// If address does not exist, add it to config.BalancesList
			if !addressExists {
				newBalance := gnoland.Balance{
					Address: callMsg.Caller,
					Amount:  defaultPremineBalance,
				}
				nodeConfig.BalancesList = append(nodeConfig.BalancesList, newBalance)
			}
		}
	}
}

// setupDevNode initializes and returns a new DevNode.
func setupDevNode(ctx context.Context, cfg *AppConfig, nodeConfig *gnodev.NodeConfig, paths ...string) (*gnodev.Node, error) {
	logger := nodeConfig.Logger

	if cfg.txsFile != "" { // Load txs files
		var err error
		nodeConfig.InitialTxs, err = gnoland.ReadGenesisTxs(ctx, cfg.txsFile)
		if err != nil {
			return nil, fmt.Errorf("unable to load transactions: %w", err)
		}

		extractDependenciesFromTxs(nodeConfig, &paths)
	} else if cfg.genesisFile != "" { // Load genesis file
		state, err := extractAppStateFromGenesisFile(cfg.genesisFile)
		if err != nil {
			return nil, fmt.Errorf("unable to load genesis file %q: %w", cfg.genesisFile, err)
		}

		// Override balances and txs
		nodeConfig.BalancesList = state.Balances

		stateTxs := state.Txs
		nodeConfig.InitialTxs = slices.Clone(stateTxs)

		logger.Info("genesis file loaded", "path", cfg.genesisFile, "txs", len(stateTxs))
	}

	if len(paths) > 0 {
		logger.Info("packages", "paths", paths)
	} else {
		logger.Debug("no path(s) provided")
	}

	return gnodev.NewDevNode(ctx, nodeConfig, paths...)
}

// setupDevNodeConfig creates and returns a new dev.NodeConfig.
func setupDevNodeConfig(
	cfg *AppConfig,
	logger *slog.Logger,
	emitter emitter.Emitter,
	balances gnoland.Balances,
	loader packages.Loader,
) *gnodev.NodeConfig {
	config := gnodev.DefaultNodeConfig(cfg.root, cfg.chainDomain)
	config.Loader = loader

	config.Logger = logger
	config.Emitter = emitter
	config.BalancesList = balances.List()
	config.TMConfig.RPC.ListenAddress = cfg.nodeRPCListenerAddr
	config.NoReplay = cfg.noReplay
	config.MaxGasPerBlock = cfg.maxGas
	config.ChainID = cfg.chainId

	// other listeners
	config.TMConfig.P2P.ListenAddress = defaultLocalAppConfig.nodeP2PListenerAddr
	config.TMConfig.ProxyApp = defaultLocalAppConfig.nodeProxyAppListenerAddr

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

func resolveUnixOrTCPAddr(in string) (addr net.Addr) {
	var err error

	if strings.HasPrefix(in, "unix://") {
		in = strings.TrimPrefix(in, "unix://")
		if addr, err = net.ResolveUnixAddr("unix", in); err == nil {
			return addr
		}

		err = fmt.Errorf("unable to resolve unix address `unix://%s`: %w", in, err)
	} else { // don't bother to checking prefix
		in = strings.TrimPrefix(in, "tcp://")
		if addr, err = net.ResolveTCPAddr("tcp", in); err == nil {
			return addr
		}

		err = fmt.Errorf("unable to resolve tcp address `tcp://%s`: %w", in, err)
	}

	panic(err)
}
