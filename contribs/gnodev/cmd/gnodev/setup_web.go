package main

import (
	"log/slog"
	"net/http"

	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
)

// setupGnowebServer initializes and starts the Gnoweb server.
func setupGnoWebServer(logger *slog.Logger, cfg *devCfg, dnode *gnodev.Node) http.Handler {
	webConfig := gnoweb.NewDefaultConfig()
	webConfig.RemoteAddr = dnode.GetRemoteAddress()
	webConfig.HelpRemote = dnode.GetRemoteAddress()
	webConfig.HelpChainID = cfg.chainId

	app := gnoweb.MakeApp(logger, webConfig)
	return app.Router
}
