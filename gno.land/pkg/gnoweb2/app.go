package gnoweb

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/service"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/yuin/goldmark"
)

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
		AssetsPath: "public",
	}
}

type App struct {
	mux     *http.ServeMux
	handler WebHandler
}

func (a *App) Router(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func MakeApp(logger *slog.Logger, cfg *AppConfig) (http.Handler, error) {
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
	cl := gnoclient.Client{
		Signer:    signer,
		RPCClient: client,
	}
	webcli := service.NewWebRender(logger, &cl, md)

	var webConfig WebHandlerConfig

	webConfig.RenderClient = webcli

	// static meta
	webConfig.Meta.AssetsPath = cfg.AssetsPath
	webConfig.Meta.RemoteHelp = cfg.RemoteHelp
	webConfig.Meta.ChaindID = cfg.ChainID

	// Setup main handler
	webhandler := NewWebHandler(logger, webConfig)

	mux := http.NewServeMux()

	// Setup Alias Middleware
	mux.Handle("/", AliasAndRedirectMiddleware(webhandler))

	// Setup asset path
	mux.Handle(cfg.AssetsPath, AssetHandler())

	return mux, nil
}

func generateWebSigner(chainid string) (gnoclient.Signer, error) {
	mnemo := "index brass unknown lecture autumn provide royal shrimp elegant wink now zebra discover swarm act ill you bullet entire outdoor tilt usage gap multiply"
	bip39Passphrase := ""
	account, index := uint32(0), uint32(0)
	return gnoclient.SignerFromBip39(mnemo, chainid, bip39Passphrase, account, index)
}
