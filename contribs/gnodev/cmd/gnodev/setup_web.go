package main

import (
	"log/slog"
	"net/http"

	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
)

// setupGnowebServer initializes and starts the Gnoweb server.
func setupGnoWebServer(logger *slog.Logger, cfg *devCfg, dnode *gnodev.Node) http.Handler {
	cfg.webConfig.RemoteAddr = dnode.GetRemoteAddress()
	if cfg.webConfig.HelpRemote == "" {
		cfg.webConfig.HelpRemote = dnode.GetRemoteAddress()
	}
	cfg.webConfig.HelpChainID = cfg.chainId

	app := gnoweb.MakeApp(logger, cfg.webConfig)
	return app.Router
}
