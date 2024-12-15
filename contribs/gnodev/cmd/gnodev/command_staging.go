package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"go.uber.org/zap/zapcore"
)

type stagingCfg struct {
	dev devCfg
}

var defaultStagingOptions = devCfg{
	chainId:             "staging",
	chainDomain:         DefaultDomain,
	maxGas:              10_000_000_000,
	webListenerAddr:     "127.0.0.1:8888",
	nodeRPCListenerAddr: "127.0.0.1:26657",
	deployKey:           DefaultDeployerAddress.String(),
	home:                gnoenv.HomeDir(),
	root:                gnoenv.RootDir(),
	interactive:         false,
	unsafeAPI:           false,
	paths:               varStrings{filepath.Join(DefaultDomain, "/**")}, // Load every package under the main domain},

	// As we have no reason to configure this yet, set this to random port
	// to avoid potential conflict with other app
	nodeP2PListenerAddr:      "tcp://127.0.0.1:0",
	nodeProxyAppListenerAddr: "tcp://127.0.0.1:0",
}

func NewStagingCmd(io commands.IO) *commands.Command {
	var cfg stagingCfg

	return commands.NewCommand(
		commands.Metadata{
			Name:          "staging",
			ShortUsage:    "gnodev staging [flags] <key-name>",
			ShortHelp:     "start gnodev in staging mode",
			NoParentFlags: true,
		},
		&cfg,
		func(_ context.Context, args []string) error {
			return execStagingCmd(&cfg, args, io)
		},
	)
}

func (c *stagingCfg) RegisterFlags(fs *flag.FlagSet) {
	c.dev.registerFlagsWithDefault(defaultStagingOptions, fs)
}

func execStagingCmd(cfg *stagingCfg, args []string, io commands.IO) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup trap signal
	osm.TrapSignal(cancel)

	level := zapcore.InfoLevel
	if cfg.dev.verbose {
		level = zapcore.DebugLevel
	}

	// Set up the logger
	logger := log.ZapLoggerToSlog(log.NewZapJSONLogger(io.Out(), level))

	// Setup trap signal
	devServer := NewApp(ctx, logger, &cfg.dev, io)
	if err := devServer.Setup(); err != nil {
		return err
	}

	return devServer.RunServer(ctx)
}

func (ds *App) RunServer(ctx context.Context) error {
	ctx, cancelWith := context.WithCancelCause(ctx)
	defer cancelWith(nil)

	addr := ds.cfg.webListenerAddr

	server := &http.Server{
		Handler:           ds.setupHandlers(),
		Addr:              ds.cfg.webListenerAddr,
		ReadHeaderTimeout: time.Second * 60,
	}

	ds.logger.WithGroup(WebLogName).Info("gnoweb started", "lisn", fmt.Sprintf("http://%s", addr))
	go func() {
		err := server.ListenAndServe()
		cancelWith(err)
	}()

	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case _, ok := <-ds.watcher.PackagesUpdate:
			if !ok {
				return nil
			}
			ds.logger.WithGroup(NodeLogName).Info("reloading...")
			if err := ds.devNode.Reload(ds.ctx); err != nil {
				ds.logger.WithGroup(NodeLogName).Error("unable to reload node", "err", err)
			}
			ds.watcher.UpdatePackagesWatch(ds.devNode.ListPkgs()...)
		}
	}
}
