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

	webConfig.HelpChainID = cfg.chainId
	webConfig.RemoteAddr = dnode.GetRemoteAddress()
	webConfig.HelpRemote = cfg.webRemoteHelperAddr

	// If `HelpRemote` is empty default it to `RemoteAddr`
	if webConfig.HelpRemote == "" {
		webConfig.HelpRemote = webConfig.RemoteAddr
	}

	app := gnoweb.MakeApp(logger, webConfig)
	return app.Router
}
