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
	"go.uber.org/zap/zapcore"
)

type webCfg struct {
	chainid    string
	remote     string
	remoteHelp string
	bind       string
}

var defaultWebOptions = &webCfg{
	chainid:    "dev",
	remote:     "127.0.0.1:26657",
	bind:       ":8888",
	remoteHelp: "",
}

func main() {
	cfg := &webCfg{}

	stdio := commands.NewDefaultIO()
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "gnoweb",
			ShortUsage: "gnoweb [flags] [path ...]",
			ShortHelp:  "runs gno.land web interface",
			LongHelp:   `gnoweb web interface`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execWeb(cfg, args, stdio)
		})

	cmd.Execute(context.Background(), os.Args[1:])
}

func (c *webCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remote,
		"remote",
		defaultWebOptions.remote,
		"target remote",
	)

	fs.StringVar(
		&c.chainid,
		"help-remote",
		defaultWebOptions.remoteHelp,
		"help page's remote address",
	)

	fs.StringVar(
		&c.chainid,
		"chain-id",
		defaultWebOptions.chainid,
		"target chain id",
	)

	fs.StringVar(
		&c.bind,
		"bind",
		defaultWebOptions.bind,
		"gnoweb listener",
	)
}

func execWeb(cfg *webCfg, args []string, io commands.IO) (err error) {
	zapLogger := log.NewZapConsoleLogger(io.Out(), zapcore.DebugLevel)
	defer zapLogger.Sync()

	// Setup logger
	logger := log.ZapLoggerToSlog(zapLogger)

	appcfg := gnoweb.NewDefaultAppConfig()
	appcfg.ChainID = cfg.chainid
	appcfg.Remote = cfg.remote
	appcfg.RemoteHelp = cfg.remoteHelp
	if appcfg.RemoteHelp == "" {
		appcfg.RemoteHelp = appcfg.Remote
	}

	app, err := gnoweb.MakeRouterApp(logger, appcfg)
	if err != nil {
		return fmt.Errorf("unable to start gnoweb app: %w", err)
	}

	bindaddr, err := net.ResolveTCPAddr("tcp", cfg.bind)
	if err != nil {
		return fmt.Errorf("unable to resolve listener %q: %w", cfg.bind, err)
	}

	logger.Info("Running", "listener", bindaddr.String())

	server := &http.Server{
		Handler:           app,
		Addr:              bindaddr.String(),
		ReadHeaderTimeout: 60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error("HTTP server stopped", " error:", err)
		os.Exit(1)
	}

	return nil
}
