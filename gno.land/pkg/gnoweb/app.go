package gnoweb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
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
	// FaucetURL is where `/faucet` redirects and the faucet the footer
	// advertises. Empty means this deployment has no faucet.
	FaucetURL string
	// StatusURL is the status page the footer links to. Empty means no link.
	StatusURL string
	// NetworkKind defaults to testnet when left empty; mainnet is explicit.
	NetworkKind components.NetworkKind
	// Domain is the domain used by the node.
	Domain string

	// CanonicalOrigin is the public origin this deployment is reachable at,
	// scheme included. Empty means no canonical tag: a canonical naming a host
	// the visitor did not reach tells a crawler the content belongs elsewhere,
	// and every deployment but one would be claiming gno.land's.
	CanonicalOrigin string
	// NoIndex keeps this deployment out of search engines (testnets, staging):
	// every page is noindex, nofollow with no canonical, and no sitemap is
	// served. Off by default, so a forgotten flag never drops a site from search.
	NoIndex bool
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
	// ?state* URLs and, in a separate bucket, $search (also used as the
	// token-bucket burst). 0 ⇒ the HTTPHandler default (1200/min).
	// ADR-003 §Resource bounds.
	StateRateLimitPerMinute int
	// StateRateLimitTrustedProxies is the list of trusted reverse-proxy
	// CIDRs (or bare IPs) for the per-IP rate limiter. X-Real-IP is honored
	// only for connections originating inside one of these networks, and so
	// is X-Forwarded-Host for the page origin (shareable links, AI prompts);
	// empty (the default) trusts nothing, so untrusted deployments never
	// trust attacker-controlled headers. ADR-003 §Resource bounds.
	StateRateLimitTrustedProxies []string
	// IndexerURL, when non-empty, is the GraphQL endpoint of a tx-indexer
	// gnoweb consults for search qualifiers the RPC node cannot answer
	// (transactions, account activity, source-wide text search).
	//
	// Empty is the default and the whole feature switch: with no indexer,
	// gnoweb makes no outbound request beyond its RPC node, and the
	// qualifiers that would need one are not registered — they do not
	// autocomplete and cannot be typed into a failing search. Indexer data is
	// never consensus data and is labelled as such wherever it appears.
	IndexerURL string
	// IndexerToken, when set, is sent as a bearer credential to the indexer.
	IndexerToken string
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
		NodeRemote:         localRemote, // local first
		RemoteHelp:         localRemote, // local first
		NodeRequestTimeout: time.Minute,
		AssetsPath:         "/public/",
		Domain:             "gno.land",
		Aliases:            maps.Clone(DefaultAliases), // callers add their own
		RenderConfig:       NewDefaultRenderConfig(),
		// StateRateLimitPerMinute stays 0 so the handler default applies.
		MaxConcurrentRPC: 32,
		TrustedPaths:     strings.Split(DefaultTrustedPaths, ","),
		IndexCommunity:   IndexNoCommunity, // names are free: see the ADR
	}
}

// chainIDRe guards every consumer at once: the value reaches markdown and a
// meta tag wallets read, and a backtick in it would close a code span.
var chainIDRe = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,64}$`)

func networkBannerText(kind components.NetworkKind, chainID string) string {
	if kind == components.NetworkLocal {
		return fmt.Sprintf("**Local development chain** — `%s`. Not a gno.land network.", chainID)
	}
	return fmt.Sprintf("**Not gno.land mainnet** — chain `%s`. Tokens and data here are not real.", chainID)
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
	logger.Info("crawl policy", "noindex", cfg.NoIndex, "canonical_origin", canonicalOrigin)

	// Initialize RPC Client.
	rpcclient, err := client.NewHTTPClient(cfg.NodeRemote,
		client.WithRequestTimeout(cfg.NodeRequestTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create HTTP client: %w", err)
	}

	// Operator-set, never guessed from the chain-id. Default is the safe kind.
	// Validated before the chain-id probe so a typo fails without a round-trip.
	if cfg.NetworkKind == "" {
		cfg.NetworkKind = components.NetworkTestnet
	}
	if !cfg.NetworkKind.Valid() {
		return nil, fmt.Errorf("invalid network kind %q, want %q, %q or %q",
			cfg.NetworkKind, components.NetworkMainnet, components.NetworkTestnet, components.NetworkLocal)
	}

	// Operator-set and rendered as a link on every page, so it must be an
	// absolute http(s) URL with no credentials; refuse to start rather than
	// publish anything else. Checked before the chain-id probe, like the kind.
	if cfg.StatusURL != "" {
		u, err := url.Parse(cfg.StatusURL)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil {
			return nil, fmt.Errorf("invalid status url %q, want an absolute http(s) URL", cfg.StatusURL)
		}
	}

	if cfg.ChainID == "" {
		cfg.ChainID, err = getChainID(context.Background(), rpcclient)
		if err != nil {
			logger.Error("unable to guess chain-id, make sure that the remote node is up and running and the RPC endpoint is valid", "error", err)
			return nil, errors.New("no chain-id configured")
		}
	}

	if !chainIDRe.MatchString(cfg.ChainID) {
		return nil, fmt.Errorf("invalid chain-id %q", cfg.ChainID)
	}

	logger.Info("network", "kind", cfg.NetworkKind, "chain-id", cfg.ChainID)

	// An operator banner wins: it has something more specific to say.
	if !cfg.Banner.Enabled() && !cfg.NetworkKind.IsMainnet() {
		banner, bannerErr := components.NewBannerData(networkBannerText(cfg.NetworkKind, cfg.ChainID), "")
		if bannerErr != nil {
			return nil, fmt.Errorf("unable to build the network banner: %w", bannerErr)
		}
		cfg.Banner = banner
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
		NetworkKind:       cfg.NetworkKind,
		FaucetURL:         cfg.FaucetURL,
		StatusURL:         cfg.StatusURL,
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
	// Built before the handler so the search feature and /search.json share
	// one directory — and therefore one singleflight group.
	searchDir := newRPCRealmDirectory(adpcli, cfg.Domain, searchMaxConcurrentQueries)

	handlerCfg := &HTTPHandlerConfig{
		ClientAdapter:                adpcli,
		Directory:                    searchDir,
		Meta:                         staticMeta,
		Renderer:                     renderer,
		Aliases:                      cfg.Aliases,
		TrustedPaths:                 cfg.TrustedPaths,
		IndexCommunity:               cfg.IndexCommunity,
		NoIndex:                      cfg.NoIndex,
		Timeout:                      cfg.NodeRequestTimeout,
		StateRateLimitPerMinute:      cfg.StateRateLimitPerMinute,
		StateRateLimitTrustedProxies: cfg.StateRateLimitTrustedProxies,
	}

	// Assigned inside the branch, never before it: an interface field holding
	// a typed nil pointer is not nil, and that is exactly how a deployment
	// with no indexer would end up advertising indexer-backed search.
	if cfg.IndexerURL != "" {
		if err := indexer.ValidateURL(cfg.IndexerURL); err != nil {
			return nil, err
		}
		// Redacted: the URL is where an indexer credential would sit.
		logger.Info("indexer enabled", "url", indexer.Redact(cfg.IndexerURL))
		handlerCfg.Indexer = indexer.New(cfg.IndexerURL, cfg.IndexerToken)
	}

	httphandler, err := NewHTTPHandler(logger, handlerCfg)
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
	mux.Handle("/search.json", handlerSearchJSON(logger, searchDir))

	// The sitemap needs an origin for absolute URLs and is never served under
	// -no-index; robots.txt names it only when it is.
	sitemapOrigin, sitemap := "", http.NotFoundHandler()
	if canonicalOrigin != "" && !cfg.NoIndex {
		// A route registered on mux shadows an alias of the same path, and
		// mux redirects a path it would clean, such as /a/../b.
		servedByHandler := func(p string) bool {
			_, pattern := mux.Handler(&http.Request{Method: http.MethodGet, URL: &url.URL{Path: p}})
			clean := path.Clean(p)
			if strings.HasSuffix(p, "/") && clean != "/" {
				clean += "/"
			}
			return pattern == "/" && clean == p
		}
		sitemapOrigin, sitemap = canonicalOrigin, handlerSitemapXML(httphandler, servedByHandler)
	}
	mux.Handle("/robots.txt", handlerRobotsTXT(cfg.NoIndex, sitemapOrigin))
	mux.Handle("/sitemap.xml", sitemap)
	if cfg.NoIndex {
		return noIndexMiddleware(mux), nil
	}

	return mux, nil
}
