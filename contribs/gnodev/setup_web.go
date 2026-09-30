package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
)

// setupGnoWebServer initializes the gnoweb HTTP handler from the gnodev
// AppConfig, returning a 404 handler when gnoweb is disabled.
func setupGnoWebServer(logger *slog.Logger, cfg *AppConfig, remoteAddr string) (http.Handler, error) {
	if cfg.noWeb {
		return http.HandlerFunc(http.NotFound), nil
	}

	appcfg := gnoweb.NewDefaultAppConfig()
	appcfg.UnsafeHTML = cfg.webHTML
	appcfg.Analytics = cfg.webAnalytics
	appcfg.AnalyticsHostname = cfg.webAnalyticsHostname
	appcfg.NodeRemote = remoteAddr
	appcfg.ChainID = cfg.chainId
	if cfg.webRemoteHelperAddr != "" {
		appcfg.RemoteHelp = cfg.webRemoteHelperAddr
	} else {
		appcfg.RemoteHelp = remoteAddr
	}

	// Staging serves its own landing page in place of the aliased
	// /r/gnoland/home, unless the operator picked a home explicitly. Rendering
	// the real gno.land homepage on a preview chain is the most effective way
	// to make it look like the live network.
	//
	// The empty/"/"/":none:" set matches how app.go reads -web-home: all three
	// mean "no explicit home". An operator who wants the old behaviour passes
	// -web-home /r/gnoland/home.
	//
	// webRemoteHelperAddr, not appcfg.RemoteHelp: the latter falls back to this
	// node's own listen address, which is useless to a remote visitor and would
	// put 127.0.0.1 into a command they are meant to copy.
	if cfg.staging {
		switch cfg.webHome {
		case "", "/", ":none:":
			appcfg.Aliases["/"] = gnoweb.AliasTarget{
				Value: stagingHomeMarkdown(cfg.chainId, cfg.webRemoteHelperAddr),
				Kind:  gnoweb.StaticMarkdown,
			}
		}
	}

	router, err := gnoweb.NewRouter(logger, appcfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create router app: %w", err)
	}

	logger.Debug("gnoweb router created",
		"remote", appcfg.NodeRemote,
		"helper_remote", appcfg.RemoteHelp,
		"html", appcfg.UnsafeHTML,
		"analytics", appcfg.Analytics,
		"chain_id", cfg.chainId,
	)
	return router, nil
}
