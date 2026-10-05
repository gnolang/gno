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
	"/":           {Value: "/r/gnoland/home", Kind: GnowebPath},
	"/about":      {Value: "/r/gnoland/pages:p/about", Kind: GnowebPath},
	"/gnolang":    {Value: "/r/gnoland/pages:p/gnolang", Kind: GnowebPath},
	"/ecosystem":  {Value: "/r/gnoland/pages:p/ecosystem", Kind: GnowebPath},
	"/start":      {Value: "/r/gnoland/pages:p/start", Kind: GnowebPath},
	"/license":    {Value: "/r/gnoland/pages:p/license", Kind: GnowebPath},
	"/contribute": {Value: "/r/gnoland/pages:p/contribute", Kind: GnowebPath},
	"/links":      {Value: "/r/gnoland/pages:p/links", Kind: GnowebPath},
	"/events":     {Value: "/r/devrels/events", Kind: GnowebPath},
	"/partners":   {Value: "/r/gnoland/pages:p/partners", Kind: GnowebPath},
	"/docs":       {Value: "/u/docs", Kind: GnowebPath},
}

// DefaultTrustedPaths are the namespaces gno.land treats as official, comma
// separated as -trusted-paths takes them: namespaces whose code the gno.land
// team reviews or whose deploy key belongs to a party it vouches for (see
// the realm notice ADR). Trust in a namespace holds only on a chain that
// enforces who may deploy under it, as r/sys/names does once enabled.
const DefaultTrustedPaths = "gnoland,sys,gov,nt,docs,demo,tests,gnops,devrels,moul,aeddi,aib,howl,leon,jeronimoalbi,mason,samcrew,onbloc,gnoswap"

// AppConfig contains configuration for gnoweb.
type AppConfig struct {
	// UnsafeHTML, if enabled, allows to use HTML in the markdown.
	UnsafeHTML bool
	// Analytics enables SimpleAnalytics.
	Analytics bool
	// AnalyticsHostname, when non-empty, is rendered as data-hostname on the
	// SimpleAnalytics script tag to override the hostname SA reports.
	// Set this when the site listens on a host SA would otherwise report
	// incorrectly (for example a non-default port in local development).
	AnalyticsHostname string
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

	// CanonicalOrigin is the public origin this deployment is reachable at,
	// scheme included. Empty means no canonical tag: a canonical naming a host
	// the visitor did not reach tells a crawler the content belongs elsewhere,
	// and every deployment but one would be claiming gno.land's.
	CanonicalOrigin string
	// Banner, if set, displays a site-wide banner above the header.
	Banner components.BannerData
	// RealmNotice, if set, is shown as the header's second row on pages of
	// packages outside TrustedPaths.
	RealmNotice components.RealmNotice
	// TrustedPaths are namespaces or package paths ("gnoland", "gnoswap/v1/pool";
	// no "/r/" or "/p/" prefix), or "*" for every path, whose pages are official: they never show
	// RealmNotice, may lend their own heading and summary to the page
	// metadata, keep their internal links followed, and are indexed whatever
	// IndexCommunity says.
	TrustedPaths []string
	// IndexCommunity says which pages outside TrustedPaths search engines may
	// index; the others get noindex, nofollow and no canonical.
	IndexCommunity CommunityIndex
	// Aliases is a map of aliases pointing to another path or a static file.
	Aliases map[string]AliasTarget
	// RenderConfig defines the default configuration for rendering realms and source files.
	RenderConfig RenderConfig
	// StateRateLimitPerMinute caps the per-IP request rate against
	// ?state* URLs (also used as the token-bucket burst). 0 ⇒ the
	// HTTPHandler default (100/min). ADR-003 §Resource bounds.
	StateRateLimitPerMinute int
	// StateRateLimitTrustedProxies is the list of trusted reverse-proxy
	// CIDRs (or bare IPs) for the per-IP rate limiter. X-Real-IP is honored
	// only for connections originating inside one of these networks, and so
	// is X-Forwarded-Host for the page origin (shareable links, AI prompts);
	// empty (the default) trusts nothing, so untrusted deployments never
	// trust attacker-controlled headers. ADR-003 §Resource bounds.
	StateRateLimitTrustedProxies []string
	// MaxConcurrentRPC caps in-flight outbound RPCs per gnoweb instance
	// against the chain node. 0 ⇒ the rpcClient default (32). Tighten on
	// chain nodes under pressure; relax when capacity allows. ADR-003
	// §Resource bounds.
	MaxConcurrentRPC int
}

// NewDefaultAppConfig returns a new default AppConfig. The default sets
// 127.0.0.1:26657 as the remote node, "dev" as the chain ID, and sets up assets
// to be served on /public/.
func NewDefaultAppConfig() *AppConfig {
	const localRemote = "127.0.0.1:26657"
	return &AppConfig{
		NodeRemote:              localRemote, // local first
		RemoteHelp:              localRemote, // local first
		NodeRequestTimeout:      time.Minute,
		AssetsPath:              "/public/",
		Domain:                  "gno.land",
		Aliases:                 DefaultAliases,
		RenderConfig:            NewDefaultRenderConfig(),
		StateRateLimitPerMinute: 100,
		MaxConcurrentRPC:        32,
		TrustedPaths:            strings.Split(DefaultTrustedPaths, ","),
		IndexCommunity:          IndexRegisteredCommunity,
	}
}

// NewRouter initializes the gnoweb router with the specified logger and configuration.
// It sets up all routes, static asset handling, and middleware.
func NewRouter(logger *slog.Logger, cfg *AppConfig) (http.Handler, error) {
	assetsBase := "/" + strings.Trim(cfg.AssetsPath, "/") + "/" // sanitize

	// A canonical is resolved against the page, so an origin without a scheme
	// would name a 404 on every page; refuse it here rather than ship it.
	canonicalOrigin, err := normalizeCanonicalOrigin(cfg.CanonicalOrigin)
	if err != nil {
		return nil, err
	}

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
	adpcli := NewRPCClientAdapter(logger, rpcclient, cfg.Domain, cfg.MaxConcurrentRPC)

	// Setup StaticMetadata
	chromaStylePath := path.Join(assetsBase, "_chroma", "style.css")

	staticMeta := StaticMetadata{
		Domain:            cfg.Domain,
		CanonicalOrigin:   canonicalOrigin,
		AssetsPath:        assetsBase,
		ChromaPath:        chromaStylePath,
		RemoteHelp:        cfg.RemoteHelp,
		ChainId:           cfg.ChainID,
		Analytics:         cfg.Analytics,
		AnalyticsHostname: cfg.AnalyticsHostname,
		AssetsVersion:     AssetsVersion(),
		Banner:            cfg.Banner,
		RealmNotice:       cfg.RealmNotice,
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
		ClientAdapter:                adpcli,
		Meta:                         staticMeta,
		Renderer:                     renderer,
		Aliases:                      cfg.Aliases,
		TrustedPaths:                 cfg.TrustedPaths,
		IndexCommunity:               cfg.IndexCommunity,
		Timeout:                      cfg.NodeRequestTimeout,
		StateRateLimitPerMinute:      cfg.StateRateLimitPerMinute,
		StateRateLimitTrustedProxies: cfg.StateRateLimitTrustedProxies,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create web handler: %w", err)
	}

	// Setup HTTP muxer
	mux := http.NewServeMux()

	// Handle web handler with redirect middleware
	mux.Handle("/", RedirectMiddleware(httphandler, staticMeta))

	// Register faucet URL to `/faucet` if specified
	if cfg.FaucetURL != "" {
		mux.Handle("/faucet", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, cfg.FaucetURL, http.StatusFound)
			components.RedirectView(components.RedirectData{
				To:        cfg.FaucetURL,
				Analytics: staticMeta.RedirectAnalytics(),
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

	// Handle realm/package discovery search (browser fetches the list once and filters locally)
	searchDir := newRPCRealmDirectory(adpcli, cfg.Domain, searchMaxConcurrentQueries)
	mux.Handle("/search.json", handlerSearchJSON(logger, searchDir))

	return mux, nil
}
