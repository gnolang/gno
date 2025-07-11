package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
)

// setupGnowebServer initializes and starts the Gnoweb server.
func setupGnoWebServer(logger *slog.Logger, cfg *AppConfig, remoteAddr string) (http.Handler, error) {
	if cfg.noWeb {
		return http.HandlerFunc(http.NotFound), nil
	}

	appcfg := gnoweb.NewDefaultAppConfig()
	appcfg.UnsafeHTML = cfg.webHTML
	appcfg.NodeRemote = remoteAddr
	appcfg.ChainID = cfg.chainId
	if cfg.webRemoteHelperAddr != "" {
		appcfg.RemoteHelp = cfg.webRemoteHelperAddr
	} else {
		appcfg.RemoteHelp = remoteAddr
	}

	router, err := gnoweb.NewRouter(logger, appcfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create router app: %w", err)
	}

	logger.Debug("gnoweb router created",
		"remote", appcfg.NodeRemote,
		"helper_remote", appcfg.RemoteHelp,
		"html", appcfg.UnsafeHTML,
		"chain_id", cfg.chainId,
	)
	return router, nil
}
