package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
)

// setupGnowebServer initializes and starts the Gnoweb server.
func setupGnoWebServer(logger *slog.Logger, cfg *devCfg, remoteAddr string) (http.Handler, error) {
	if cfg.noWeb {
		return http.HandlerFunc(http.NotFound), nil
	}

	fmt.Printf("REMOTE: %+v\r\n", remoteAddr)
	appcfg := gnoweb.NewDefaultAppConfig()
	appcfg.UnsafeHTML = cfg.webHTML
	appcfg.NodeRemote = remoteAddr
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
