package main

import (
	"encoding/json"
	"fmt"
	"github.com/gnolang/gno/contribs/gnoserve/pkg/handler"
	"github.com/gnolang/gno/contribs/gnoserve/pkg/markdown"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/yuin/goldmark"
	mdhtml "github.com/yuin/goldmark/renderer/html"
	"log/slog"
	"net/http"
	"path"
	"strings"
)

// FIXME: hacked up from gnoweb - make this useful again
func handlerStatusJSON(logger *slog.Logger, cli *client.RPCClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		qres, err := cli.ABCIQuery(".app/version", []byte{})
		if err != nil {
			logger.Error("query app version", "error", err)
			http.Error(w, "query app version: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(qres.Response); err != nil {
			logger.Error("encode response", "error", err)
			http.Error(w, "encode response: "+err.Error(), http.StatusInternalServerError)
		}
	})
}

type GnomarkExt interface {
	Extend(goldmark.Markdown)
}

// NewHTMLClient creates a new instance of WebClient.
// it registers the GnoMark markdown extension
func NewHTMLClient(log *slog.Logger, cfg *gnoweb.HTMLWebClientConfig) (client *gnoweb.HTMLWebClient) {
	client = gnoweb.NewHTMLClient(log, cfg)
	NewGnomarkExt(client).Extend(client.Markdown)
	return client
}

func NewGnomarkExt(client *gnoweb.HTMLWebClient) GnomarkExt {
	return &markdown.GnoMarkExtension{
		Client: client,
	}
}

func NewRouter(logger *slog.Logger, cfg *gnoweb.AppConfig) (http.Handler, error) {
	// Initialize RPC Client
	httpClient, err := client.NewHTTPClient(cfg.NodeRemote)
	if err != nil {
		return nil, fmt.Errorf("unable to create HTTP httpClient: %w", err)
	}

	// Setup web httpClient HTML
	webcfg := gnoweb.NewDefaultHTMLWebClientConfig(httpClient)
	webcfg.Domain = cfg.Domain
	if cfg.UnsafeHTML {
		webcfg.GoldmarkOptions = append(webcfg.GoldmarkOptions, goldmark.WithRendererOptions(
			mdhtml.WithXHTML(), mdhtml.WithUnsafe(),
		))
	}
	webcli := NewHTMLClient(logger, webcfg)

	// Setup StaticMetadata
	chromaStylePath := path.Join(cfg.AssetsPath, "_chroma", "style.css")
	staticMeta := handler.StaticMetadata{
		Domain:     cfg.Domain,
		AssetsPath: cfg.AssetsPath,
		ChromaPath: chromaStylePath,
		RemoteHelp: cfg.RemoteHelp,
		ChainId:    cfg.ChainID,
		Analytics:  cfg.Analytics,
	}

	// Configure WebHandler
	webConfig := handler.WebHandlerConfig{WebClient: webcli, Meta: staticMeta}
	webhandler, err := handler.NewWebHandler(logger, webConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create web handler: %w", err)
	}

	// Setup HTTP muxer
	mux := http.NewServeMux()

	// Handle web handler with alias middleware
	mux.Handle("/", gnoweb.AliasAndRedirectMiddleware(webhandler, cfg.Analytics))

	// Register faucet URL to `/faucet` if specified
	if cfg.FaucetURL != "" {
		mux.Handle("/faucet", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, cfg.FaucetURL, http.StatusFound)
			_ = components.RedirectView(components.RedirectData{
				To:            cfg.FaucetURL,
				WithAnalytics: cfg.Analytics,
			}).Render(w)
		}))
	}

	// Handle Chroma CSS requests
	// XXX: probably move this elsewhere
	mux.Handle(chromaStylePath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		if err := webcli.WriteFormatterCSS(w); err != nil {
			logger.Error("unable to write CSS", "err", err)
			http.NotFound(w, r)
		}
	}))

	// Handle assets path
	// XXX: add caching
	assetsBase := "/" + strings.Trim(cfg.AssetsPath, "/") + "/"
	cfg.AssetsDir = "./public"
	if cfg.AssetsDir != "" {
		logger.Debug("using assets dir instead of embedded assets", "dir", cfg.AssetsDir)
		mux.Handle(assetsBase, gnoweb.DevAssetHandler(assetsBase, cfg.AssetsDir))
	} else {
		mux.Handle(assetsBase, gnoweb.AssetHandler())
	}

	// Handle status page
	mux.Handle("/status.json", handlerStatusJSON(logger, httpClient))

	return mux, nil
}

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

	router, err := NewRouter(logger, appcfg)
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
