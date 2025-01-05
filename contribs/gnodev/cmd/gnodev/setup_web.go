package main

import (
	"log/slog"
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
)

// setupGnowebServer initializes and starts the Gnoweb server.
func setupGnoWebServer(logger *slog.Logger, cfg *devCfg, remoteAddr string) http.Handler {
	webConfig := gnoweb.NewDefaultConfig()

	webConfig.HelpChainID = cfg.chainId
	webConfig.RemoteAddr = remoteAddr
	webConfig.HelpRemote = cfg.webRemoteHelperAddr
	webConfig.WithHTML = cfg.webWithHTML

	// If `HelpRemote` is empty default it to `RemoteAddr`
	if webConfig.HelpRemote == "" {
		webConfig.HelpRemote = webConfig.RemoteAddr
	}

	app := gnoweb.MakeApp(logger, webConfig)
	return app.Router
}
