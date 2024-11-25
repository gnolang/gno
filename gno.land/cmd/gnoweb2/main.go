package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	gnoweb "github.com/gnolang/gno/gno.land/pkg/gnoweb2"
	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/gno.land/pkg/service"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/yuin/goldmark"
	"go.uber.org/zap/zapcore"
)

type webCfg struct {
	chainid string
	remote  string
	bind    string
}

var defaultWebOptions = &webCfg{
	chainid: "dev",
	remote:  "127.0.0.1:26657",
	bind:    ":8888",
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
		"user's local directory for keys",
	)

	fs.StringVar(
		&c.chainid,
		"chainid",
		defaultWebOptions.chainid,
		"user's local directory for keys",
	)

	fs.StringVar(
		&c.bind,
		"bind",
		defaultWebOptions.bind,
		"user's local directory for keys",
	)

}

func execWeb(cfg *webCfg, args []string, io commands.IO) (err error) {
	zapLogger := log.NewZapConsoleLogger(os.Stdout, zapcore.DebugLevel)
	defer zapLogger.Sync()

	// Setup logger
	logger := log.ZapLoggerToSlog(zapLogger)

	md := goldmark.New()

	staticMeta := gnoweb.StaticMetadata{
		AssetsPath: "/public/",
		RemoteHelp: cfg.remote,
	}

	mux := http.NewServeMux()

	// Setup asset handler
	// if cfg.dev {
	// 	mux.Handle(staticMeta.AssetsPath, AssetDevHandler())
	// } else {
	mux.Handle(staticMeta.AssetsPath, gnoweb.AssetHandler())
	// }

	client, err := client.NewHTTPClient(cfg.remote)
	if err != nil {
		return fmt.Errorf("unable to create http client: %W", err)
	}

	mnemo := "index brass unknown lecture autumn provide royal shrimp elegant wink now zebra discover swarm act ill you bullet entire outdoor tilt usage gap multiply"
	bip39Passphrase := ""
	account, index := uint32(0), uint32(0)
	chainID := cfg.chainid
	signer, err := gnoclient.SignerFromBip39(mnemo, chainID, bip39Passphrase, account, index)
	if err != nil {
		return fmt.Errorf("unable to create signer: %w", err)
	}

	// Setup webservice
	cl := gnoclient.Client{
		Signer:    signer,
		RPCClient: client,
	}
	webcli := service.NewWebRender(logger, &cl, md)

	if len(args) > 0 {
		var qargs string
		if len(args) > 1 {
			qargs = strings.Join(args[1:], ",")
		}

		_, err = webcli.Render(io.Out(), args[0], qargs)
		return
	}

	webcfg := gnoweb.WebHandlerConfig{
		RenderClient: webcli,
		Meta:         staticMeta,
	}

	// Setup main handler
	webhandler := gnoweb.NewWebHandler(
		logger,
		webcfg,
	)

	// Setup Alias Middleware
	mux.Handle("/", gnoweb.AliasAndRedirectMiddleware(webhandler))

	bindaddr, err := net.ResolveTCPAddr("tcp", cfg.bind)
	if err != nil {
		return fmt.Errorf("unable to resolve listener: %q", cfg.bind)
	}

	logger.Info("Running", "listener", bindaddr.String())

	server := &http.Server{
		Handler:           mux,
		Addr:              bindaddr.String(),
		ReadHeaderTimeout: 60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error("HTTP server stopped", " error:", err)
		os.Exit(1)
	}

	return nil
}
