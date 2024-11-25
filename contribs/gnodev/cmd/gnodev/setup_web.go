package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	gnoweb "github.com/gnolang/gno/gno.land/pkg/gnoweb2"
	"github.com/gnolang/gno/gno.land/pkg/service"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/yuin/goldmark"
)

func makeWebApp(logger *slog.Logger, cfg *devCfg, dnode *gnodev.Node) error {

	// bindaddr, err := net.ResolveTCPAddr("tcp", cfg.bind)
	// if err != nil {
	// 	return fmt.Errorf("unable to resolve listener: %q", cfg.bind)
	// }

	// logger.Info("Running", "listener", bindaddr.String())

	// server := &http.Server{
	// 	Handler:           mux,
	// 	Addr:              bindaddr.String(),
	// 	ReadHeaderTimeout: 60 * time.Second,
	// }

	// if err := server.ListenAndServe(); err != nil {
	// 	logger.Error("HTTP server stopped", " error:", err)
	// 	os.Exit(1)
	// }

}

// setupGnowebServer initializes and starts the Gnoweb server.
func setupGnoWebServer(logger *slog.Logger, cfg *devCfg, dnode *gnodev.Node) (http.Handler, error) {
	remote := dnode.GetRemoteAddress()

	md := goldmark.New()

	client, err := client.NewHTTPClient(remote)
	if err != nil {
		return nil, fmt.Errorf("unable to create http client: %W", err)
	}

	mnemo := "index brass unknown lecture autumn provide royal shrimp elegant wink now zebra discover swarm act ill you bullet entire outdoor tilt usage gap multiply"
	bip39Passphrase := ""
	account, index := uint32(0), uint32(0)
	chainID := cfg.chainId
	signer, err := gnoclient.SignerFromBip39(mnemo, chainID, bip39Passphrase, account, index)
	if err != nil {
		return nil, fmt.Errorf("unable to create signer: %w", err)
	}

	// Setup webservice
	cl := gnoclient.Client{
		Signer:    signer,
		RPCClient: client,
	}
	webcli := service.NewWebRender(logger, &cl, md)

	var webConfig gnoweb.WebHandlerConfig

	webConfig.RenderClient = webcli

	// static meta
	// webConfig.Meta.AssetsPath = // XXX
	webConfig.Meta.RemoteHelp = cfg.webRemoteHelperAddr
	webConfig.Meta.ChaindID = cfg.chainId

	// Setup main handler
	webhandler := gnoweb.NewWebHandler(
		context.TODO(), // XXX
		logger,
		webConfig,
	)

	mux := http.NewServeMux()
	mux.Handle(webConfig.Meta.AssetsPath, gnoweb.AssetHandler())

	// Setup Alias Middleware
	mux.Handle("/", gnoweb.AliasAndRedirectMiddleware(webhandler))

	return mux, nil
}
