package gnoweb

import (
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/yuin/goldmark"
)

var chromaStyle = styles.Get("friendly")

func init() {
	if chromaStyle == nil {
		panic("unable to get chroma style")
	}
}

type AppConfig struct {
	Remote     string
	RemoteHelp string
	ChainID    string
	AssetsPath string
}

func NewDefaultAppConfig() *AppConfig {
	const defaultRemote = "127.0.0.1:26657"

	return &AppConfig{
		Remote: defaultRemote, RemoteHelp: defaultRemote, // same as `Remote` by default
		ChainID:    "dev",
		AssetsPath: "/public/",
	}
}

func MakeRouterApp(logger *slog.Logger, cfg *AppConfig) (http.Handler, error) {
	md := goldmark.New()

	client, err := client.NewHTTPClient(cfg.Remote)
	if err != nil {
		return nil, fmt.Errorf("unable to create http client: %W", err)
	}

	signer, err := generateWebSigner(cfg.ChainID)
	if err != nil {
		return nil, fmt.Errorf("unable to generate web signer: %w", err)
	}

	// Setup webservice
	gnocli := gnoclient.Client{
		Signer:    signer,
		RPCClient: client,
	}
	webcli := NewWebClient(logger, &gnocli, md)

	formatter := html.New(
		html.WithLineNumbers(true),
		html.WithClasses(true),
		html.ClassPrefix("chroma-"),
	)
	chromaStylePath := path.Join(cfg.AssetsPath, "_chroma", "style.css")

	var webConfig WebHandlerConfig

	webConfig.RenderClient = webcli
	webConfig.Formatter = newFormaterWithStyle(formatter, chromaStyle)

	// static meta
	webConfig.Meta.AssetsPath = cfg.AssetsPath
	webConfig.Meta.ChromaPath = chromaStylePath
	webConfig.Meta.RemoteHelp = cfg.RemoteHelp
	webConfig.Meta.ChaindID = cfg.ChainID

	// Setup main handler
	webhandler := NewWebHandler(logger, webConfig)

	mux := http.NewServeMux()

	// Setup Webahndler along Alias Middleware
	mux.Handle("/", AliasAndRedirectMiddleware(webhandler))

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

	mux.Handle("/status.json", handlerStatusJSON(logger, &gnocli))

	return mux, nil
}

func generateWebSigner(chainid string) (gnoclient.Signer, error) {
	mnemo := "index brass unknown lecture autumn provide royal shrimp elegant wink now zebra discover swarm act ill you bullet entire outdoor tilt usage gap multiply"
	bip39Passphrase := ""
	account, index := uint32(0), uint32(0)
	return gnoclient.SignerFromBip39(mnemo, chainid, bip39Passphrase, account, index)
}
