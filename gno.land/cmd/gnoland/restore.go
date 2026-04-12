package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/bft/backup"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/events"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/telemetry"
)

type restoreCfg struct {
	startCfg
	backupDir        string
	endHeight        int64
	skipVerification bool
}

func newRestoreCmd(io commands.IO) *commands.Command {
	cfg := &restoreCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "restore",
			ShortUsage: "restore [flags]",
			ShortHelp:  "restore the Gnoland blockchain node",
			LongHelp:   "Restores the Gnoland blockchain node, with accompanying setup",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execRestore(ctx, cfg, io)
		},
	)
}

func (c *restoreCfg) RegisterFlags(fs *flag.FlagSet) {
	c.startCfg.RegisterFlags(fs)

	fs.StringVar(
		&c.backupDir,
		"backup-dir",
		"blocks-backup",
		"directory where the backup files are",
	)

	fs.Int64Var(
		&c.endHeight,
		"end-height",
		0,
		"height at which the restore process should stop",
	)

	fs.BoolVar(
		&c.skipVerification,
		"skip-verification",
		false,
		"skip commit verification of the backup files",
	)
}

func execRestore(ctx context.Context, c *restoreCfg, io commands.IO) error {
	gnoNode, err := createNode(ctx, &c.startCfg, io)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := gnoNode.Config().LocalApp.Close(); closeErr != nil {
			io.ErrPrintln("unable to close gnoland application:", closeErr)
		}
	}()

	// need block n+1 to commit block n
	endHeight := c.endHeight
	if endHeight != 0 {
		endHeight += 1
	}

	startHeight := gnoNode.BlockStore().Height() + 1
	if c.endHeight != 0 && c.endHeight < startHeight {
		return fmt.Errorf("invalid input: requested end height (#%d) is smaller than next chain height (#%d)", c.endHeight, startHeight)
	}

	return backup.WithReader(c.backupDir, startHeight, endHeight, func(reader backup.Reader) error {
		return gnoNode.Restore(ctx, reader, c.skipVerification)
	})
}

func createNode(ctx context.Context, c *startCfg, io commands.IO) (*node.Node, error) {
	// Get the absolute path to the node's data directory
	nodeDir, err := filepath.Abs(c.dataDir)
	if err != nil {
		return nil, fmt.Errorf("unable to get absolute path for data directory, %w", err)
	}

	// Get the absolute path to the node's genesis.json
	genesisPath, err := filepath.Abs(c.genesisFile)
	if err != nil {
		return nil, fmt.Errorf("unable to get absolute path for the genesis.json, %w", err)
	}

	// Initialize the logger
	zapLogger, err := log.InitializeZapLogger(io.Out(), c.logLevel, c.logFormat)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize zap logger, %w", err)
	}

	// Wrap the zap logger
	logger := log.ZapLoggerToSlog(zapLogger)

	if c.lazyInit {
		if err := lazyInitNodeDir(io, nodeDir); err != nil {
			return nil, fmt.Errorf("unable to lazy-init the node directory, %w", err)
		}
	}

	// Load the configuration
	cfg, err := config.LoadConfig(nodeDir)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", tryConfigInit, err)
	}

	// Check if the genesis.json exists
	if !osm.FileExists(genesisPath) {
		if !c.lazyInit {
			return nil, errMissingGenesis
		}

		// Get the node key for signer init
		nodeKey, err := p2pTypes.LoadOrMakeNodeKey(cfg.NodeKeyFile())
		if err != nil {
			return nil, fmt.Errorf("unable to load or make node key, %w", err)
		}

		// Init the signer based on the config
		signer, err := privval.NewSignerFromConfig(ctx, cfg.Consensus.PrivValidator, nodeKey.PrivKey, logger)
		if err != nil {
			return nil, fmt.Errorf("unable to instantiate signer based on config: %w", err)
		}

		// Init a new genesis.json
		if err := lazyInitGenesis(io, c, genesisPath, signer); err != nil {
			return nil, fmt.Errorf("unable to initialize genesis.json, %w", err)
		}
	}

	// Initialize telemetry
	if err := telemetry.Init(*cfg.Telemetry); err != nil {
		return nil, fmt.Errorf("unable to initialize telemetry, %w", err)
	}

	// Print the starting graphic
	if c.logFormat != string(log.JSONFormat) {
		io.Println(startGraphic)
	}

	// Create a top-level shared event switch
	evsw := events.NewEventSwitch()

	// Create application and node
	cfg.LocalApp, err = gnoland.NewApp(
		nodeDir,
		gnoland.GenesisAppConfig{
			SkipFailingTxs:      c.skipFailingGenesisTxs,
			SkipSigVerification: c.skipGenesisSigVerification,
		},
		cfg.Application,
		evsw,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create the Gnoland app, %w", err)
	}

	// Create a default node, with the given setup
	opts := []node.Option{}
	if c.earlyStart {
		opts = append(opts, node.WithEarlyStart())
	}
	gnoNode, err := node.DefaultNewNode(cfg, genesisPath, evsw, logger, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to create the Gnoland node, %w", err)
	}

	return gnoNode, err
}
