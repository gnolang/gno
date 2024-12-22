package main

import (
	"fmt"
	"log/slog"
	"net/http"

	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
)

// setupGnowebServer initializes and starts the Gnoweb server.
func setupGnoWebServer(logger *slog.Logger, cfg *devCfg, dnode *gnodev.Node) (http.Handler, error) {
	if cfg.noWeb {
		return http.HandlerFunc(http.NotFound), nil
	}

	remote := dnode.GetRemoteAddress()

	appcfg := gnoweb.NewDefaultAppConfig()
	appcfg.UnsafeHTML = cfg.webHTML
	appcfg.NodeRemote = remote
	appcfg.ChainID = cfg.chainId
	if cfg.webRemoteHelperAddr != "" {
		appcfg.RemoteHelp = cfg.webRemoteHelperAddr
	}

	router, err := gnoweb.NewRouter(logger, appcfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create router app: %w", err)
	}

	return router, nil
}
