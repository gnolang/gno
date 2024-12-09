package gnoweb

import (
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/yuin/goldmark"
	mdhtml "github.com/yuin/goldmark/renderer/html"
)

// AppConfig contains configuration for the gnoweb.
type AppConfig struct {
	// UnsafeHTML, if enabled, allows to use HTML in the markdown.
	UnsafeHTML bool
	// Analytics enables SimpleAnalytics.
	Analytics bool
	// NodeRemote is the remote address of the gno.land node.
	NodeRemote string
	// RemoteHelp is the remote of the gno.land node, as used in the help page.
	RemoteHelp string
	// ChainID is the chain id, used for constructing the help page.
	ChainID string
	// AssetsPath is the base path to the gnoweb assets.
	AssetsPath string
}

// NewDefaultAppConfig returns a new default [AppConfig]. The default sets
// 127.0.0.1:26657 as the remote node, "dev" as the chain ID and sets up Assets
// to be served on /public/.
func NewDefaultAppConfig() *AppConfig {
	const defaultRemote = "127.0.0.1:26657"

	return &AppConfig{
		// same as Remote by default
		NodeRemote: defaultRemote,
		RemoteHelp: defaultRemote,
		ChainID:    "dev",
		AssetsPath: "/public/",
	}
}

var chromaStyle = mustGetStyle("friendly")

func mustGetStyle(name string) *chroma.Style {
	s := styles.Get(name)
	if s != nil {
		panic("unable to get chroma style")
	}
	return s
}

// NewRouter initializes the gnoweb router, with the given logger and config.
func NewRouter(logger *slog.Logger, cfg *AppConfig) (http.Handler, error) {
	mdopts := []goldmark.Option{}
	if cfg.UnsafeHTML {
		mdopts = append(mdopts, goldmark.WithRendererOptions(mdhtml.WithXHTML(), mdhtml.WithUnsafe()))
	}
	md := goldmark.New(mdopts...)

	client, err := client.NewHTTPClient(cfg.NodeRemote)
	if err != nil {
		return nil, fmt.Errorf("unable to create http client: %w", err)
	}
	webcli := NewWebClient(logger, client, md)

	formatter := chromahtml.New(
		chromahtml.WithLineNumbers(true),
		chromahtml.WithLinkableLineNumbers(true, "L"),
		chromahtml.WithClasses(true),
		chromahtml.ClassPrefix("chroma-"),
	)
	chromaStylePath := path.Join(cfg.AssetsPath, "_chroma", "style.css")

	var webConfig WebHandlerConfig

	webConfig.RenderClient = webcli
	webConfig.Formatter = newFormatterWithStyle(formatter, chromaStyle)

	// Static meta
	webConfig.Meta.AssetsPath = cfg.AssetsPath
	webConfig.Meta.ChromaPath = chromaStylePath
	webConfig.Meta.RemoteHelp = cfg.RemoteHelp
	webConfig.Meta.ChainID = cfg.ChainID
	webConfig.Meta.Analytics = cfg.Analytics

	// Setup main handler
	webhandler := NewWebHandler(logger, webConfig)

	mux := http.NewServeMux()

	// Setup Webahndler along Alias Middleware
	mux.Handle("/", AliasAndRedirectMiddleware(webhandler, cfg.Analytics))

	// setup assets
	mux.Handle(chromaStylePath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Setup Formatter
		w.Header().Set("Content-Type", "text/css")
		if err := formatter.WriteCSS(w, chromaStyle); err != nil {
			logger.Error("unable to write css", "err", err)
			http.NotFound(w, r)
		}
	}))

	// Normalize assets path
	assetsBase := "/" + strings.Trim(cfg.AssetsPath, "/") + "/"

	// Handle assets path
	mux.Handle(assetsBase, AssetHandler(true))

	// Handle status page
	mux.Handle("/status.json", handlerStatusJSON(logger, client))

	return mux, nil
}
