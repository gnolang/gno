package gnoweb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/yuin/goldmark"
	mdhtml "github.com/yuin/goldmark/renderer/html"
)

var DefaultAliases = map[string]AliasTarget{
	"/":           {"/r/gnoland/home", GnowebPath},
	"/about":      {"/r/gnoland/pages:p/about", GnowebPath},
	"/gnolang":    {"/r/gnoland/pages:p/gnolang", GnowebPath},
	"/ecosystem":  {"/r/gnoland/pages:p/ecosystem", GnowebPath},
	"/start":      {"/r/gnoland/pages:p/start", GnowebPath},
	"/license":    {"/r/gnoland/pages:p/license", GnowebPath},
	"/contribute": {"/r/gnoland/pages:p/contribute", GnowebPath},
	"/links":      {"/r/gnoland/pages:p/links", GnowebPath},
	"/events":     {"/r/gnoland/events", GnowebPath},
	"/partners":   {"/r/gnoland/pages:p/partners", GnowebPath},
	"/docs":       {"/u/docs", GnowebPath},
}

// AppConfig contains configuration for gnoweb.
type AppConfig struct {
	// UnsafeHTML, if enabled, allows to use HTML in the markdown.
	UnsafeHTML bool
	// Analytics enables SimpleAnalytics.
	Analytics bool
	// NodeRemote is the remote address of the gno.land node.
	NodeRemote string
	// NodeRequestTimeout define how much time a request to the remote node should live before timeout.
	NodeRequestTimeout time.Duration
	// RemoteHelp is the remote of the gno.land node, as used in the help page.
	RemoteHelp string
	// AssetsPath is the base path to the gnoweb assets.
	AssetsPath string
	// NoAssetsCache disables assets caching.
	NoAssetsCache bool
	// ChainID is the chain id, used for constructing the help page.
	ChainID string
	// FaucetURL, if specified, will be the URL to which `/faucet` redirects.
	FaucetURL string
	// Domain is the domain used by the node.
	Domain string
	// Aliases is a map of aliases pointing to another path or a static file.
	Aliases map[string]AliasTarget
	// RenderConfig defines the default configuration for rendering realms and source files.
	RenderConfig RenderConfig
}

// NewDefaultAppConfig returns a new default AppConfig. The default sets
// 127.0.0.1:26657 as the remote node, "dev" as the chain ID, and sets up assets
// to be served on /public/.
func NewDefaultAppConfig() *AppConfig {
	const localRemote = "127.0.0.1:26657"
	return &AppConfig{
		NodeRemote:         localRemote, // local first
		RemoteHelp:         localRemote, // local first
		NodeRequestTimeout: time.Minute,
		AssetsPath:         "/public/",
		Domain:             "gno.land",
		Aliases:            DefaultAliases,
		RenderConfig:       NewDefaultRenderConfig(),
	}
}

// NewRouter initializes the gnoweb router with the specified logger and configuration.
// It sets up all routes, static asset handling, and middleware.
func NewRouter(logger *slog.Logger, cfg *AppConfig) (http.Handler, error) {
	assetsBase := "/" + strings.Trim(cfg.AssetsPath, "/") + "/" // sanitize

	// Initialize RPC Client.
	rpcclient, err := client.NewHTTPClient(cfg.NodeRemote,
		client.WithRequestTimeout(cfg.NodeRequestTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create HTTP client: %w", err)
	}

	if cfg.ChainID == "" {
		cfg.ChainID, err = getChainID(context.Background(), rpcclient)
		if err != nil {
			logger.Error("unable to guess chain-id, make sure that the remote node is up and running and the RPC endpoint is valid", "error", err)
			return nil, errors.New("no chain-id configured")
		}
	}

	// Setup client adapter
	adpcli := NewRPCClientAdapter(logger, rpcclient, cfg.Domain)

	// Setup StaticMetadata
	chromaStylePath := path.Join(assetsBase, "_chroma", "style.css")

	// Build time for cache busting
	buildTime := time.Now().Format("20060102150405") // YYYYMMDDHHMMSS

	staticMeta := StaticMetadata{
		Domain:     cfg.Domain,
		AssetsPath: assetsBase,
		ChromaPath: chromaStylePath,
		RemoteHelp: cfg.RemoteHelp,
		ChainId:    cfg.ChainID,
		Analytics:  cfg.Analytics,
		BuildTime:  buildTime,
	}

	// Configure Markdown renderer
	rcfg := cfg.RenderConfig
	if cfg.UnsafeHTML {
		rcfg.GoldmarkOptions = append(rcfg.GoldmarkOptions, goldmark.WithRendererOptions(
			mdhtml.WithXHTML(), mdhtml.WithUnsafe(),
		))
	}
	renderer := NewHTMLRenderer(logger, rcfg, adpcli)

	// Configure HTTPHandler
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]AliasTarget) // Sanitize Aliases cfg
	}
	httphandler, err := NewHTTPHandler(logger, &HTTPHandlerConfig{
		ClientAdapter: adpcli,
		Meta:          staticMeta,
		Renderer:      renderer,
		Aliases:       cfg.Aliases,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create web handler: %w", err)
	}

	// Setup HTTP muxer
	mux := http.NewServeMux()

	// Handle web handler with redirect middleware
	mux.Handle("/", RedirectMiddleware(httphandler, cfg.Analytics))

	// Register faucet URL to `/faucet` if specified
	if cfg.FaucetURL != "" {
		mux.Handle("/faucet", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, cfg.FaucetURL, http.StatusFound)
			components.RedirectView(components.RedirectData{
				To:            cfg.FaucetURL,
				WithAnalytics: cfg.Analytics,
			}).Render(w)
		}))
	}

	cacheAssetHandler := DefaultCacheAssetsHandler
	if cfg.NoAssetsCache {
		cacheAssetHandler = NoCacheHandler
	}

	// Handle Chroma CSS requests
	// XXX: probably move this elsewhere
	chromaStyleHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		if err := renderer.WriteChromaCSS(w); err != nil {
			logger.Error("unable to write CSS", "err", err)
			http.NotFound(w, r)
		}
	})
	mux.Handle(chromaStylePath, cacheAssetHandler(chromaStyleHandler))

	// Handle assets path
	assetsHandler := cacheAssetHandler(AssetHandler())
	mux.Handle(assetsBase, http.StripPrefix(assetsBase, assetsHandler))

	// Handle status page
	mux.Handle("/status.json", handlerStatusJSON(logger, rpcclient))

	// Handle liveness check - service itself is up and running
	mux.Handle("/liveness", handlerLivenessJSON(logger))

	// Handle readiness check - service can communicate with RPC node and serve clients
	mux.Handle("/ready", handlerReadyJSON(logger, rpcclient, cfg.Domain))

	return mux, nil
}
