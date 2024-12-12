package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type webCfg struct {
	chainid    string
	remote     string
	remoteHelp string
	bind       string
	analytics  bool
	json       bool
	html       bool
}

var defaultWebOptions = webCfg{
	chainid: "dev",
	remote:  "127.0.0.1:26657",
	bind:    ":8888",
}

func main() {
	var cfg webCfg

	stdio := commands.NewDefaultIO()
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "gnoweb",
			ShortUsage: "gnoweb [flags] [path ...]",
			ShortHelp:  "runs gno.land web interface",
			LongHelp:   `gnoweb web interface`,
		},
		&cfg,
		func(ctx context.Context, args []string) error {
			run, err := setupWeb(&cfg, args, stdio)
			if err != nil {
				return err
			}

			return run()
		})

	cmd.Execute(context.Background(), os.Args[1:])
}

func (c *webCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remote,
		"remote",
		defaultWebOptions.remote,
		"remote gno.land node address",
	)

	fs.StringVar(
		&c.chainid,
		"help-remote",
		defaultWebOptions.remoteHelp,
		"help page's remote address",
	)

	fs.StringVar(
		&c.chainid,
		"help-chainid",
		defaultWebOptions.chainid,
		"Deprecated: use `chainid` instead",
	)

	fs.StringVar(
		&c.chainid,
		"chainid",
		defaultWebOptions.chainid,
		"target chain id",
	)

	fs.StringVar(
		&c.bind,
		"bind",
		defaultWebOptions.bind,
		"gnoweb listener",
	)

	fs.BoolVar(
		&c.json,
		"json",
		defaultWebOptions.json,
		"display log in json format",
	)

	fs.BoolVar(
		&c.html,
		"html",
		defaultWebOptions.html,
		"enable unsafe html",
	)

	fs.BoolVar(
		&c.analytics,
		"with-analytics",
		defaultWebOptions.analytics,
		"nable privacy-first analytics",
	)
}

func setupWeb(cfg *webCfg, args []string, io commands.IO) (func() error, error) {
	// Setup logger
	var zapLogger *zap.Logger
	if cfg.json {
		zapLogger = log.NewZapJSONLogger(io.Out(), zapcore.DebugLevel)
	} else {
		zapLogger = log.NewZapConsoleLogger(io.Out(), zapcore.DebugLevel)
	}
	defer zapLogger.Sync()

	logger := log.ZapLoggerToSlog(zapLogger)

	appcfg := gnoweb.NewDefaultAppConfig()
	appcfg.ChainID = cfg.chainid
	appcfg.NodeRemote = cfg.remote
	appcfg.RemoteHelp = cfg.remoteHelp
	appcfg.Analytics = cfg.analytics
	appcfg.UnsafeHTML = cfg.html
	if appcfg.RemoteHelp == "" {
		appcfg.RemoteHelp = appcfg.NodeRemote
	}

	app, err := gnoweb.NewRouter(logger, appcfg)
	if err != nil {
		return nil, fmt.Errorf("unable to start gnoweb app: %w", err)
	}

	bindaddr, err := net.ResolveTCPAddr("tcp", cfg.bind)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve listener %q: %w", cfg.bind, err)
	}

	logger.Info("Running", "listener", bindaddr.String())

	server := &http.Server{
		Handler:           app,
		Addr:              bindaddr.String(),
		ReadHeaderTimeout: 60 * time.Second,
	}

	return func() error {
		if err := server.ListenAndServe(); err != nil {
			logger.Error("HTTP server stopped", " error:", err)
			return commands.ExitCodeError(1)
		}

		return nil
	}, nil
}
