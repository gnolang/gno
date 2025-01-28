package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type webCfg struct {
	chainid    string
	remote     string
	remoteHelp string
	bind       string
	faucetURL  string
	assetsDir  string
	timeout    time.Duration
	analytics  bool
	json       bool
	html       bool
	noStrict   bool
	verbose    bool
}

var defaultWebOptions = webCfg{
	chainid: "dev",
	remote:  "127.0.0.1:26657",
	bind:    ":8888",
	timeout: time.Minute,
}

func main() {
	var cfg webCfg

	stdio := commands.NewDefaultIO()
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "gnoweb",
			ShortUsage: "gnoweb [flags] [path ...]",
			ShortHelp:  "runs gno.land web interface",
			LongHelp:   `gnoweb web interface`,
		},
		&cfg,
		func(ctx context.Context, args []string) error {
			run, err := setupWeb(&cfg, args, stdio)
			if err != nil {
				return err
			}

			return run()
		})

	cmd.Execute(context.Background(), os.Args[1:])
}

func (c *webCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remote,
		"remote",
		defaultWebOptions.remote,
		"remote gno.land node address",
	)

	fs.StringVar(
		&c.remoteHelp,
		"help-remote",
		defaultWebOptions.remoteHelp,
		"help page's remote address",
	)

	fs.StringVar(
		&c.assetsDir,
		"assets-dir",
		defaultWebOptions.assetsDir,
		"if not empty, will be use as assets directory",
	)

	fs.StringVar(
		&c.chainid,
		"help-chainid",
		defaultWebOptions.chainid,
		"Deprecated: use `chainid` instead",
	)

	fs.StringVar(
		&c.chainid,
		"chainid",
		defaultWebOptions.chainid,
		"target chain id",
	)

	fs.StringVar(
		&c.bind,
		"bind",
		defaultWebOptions.bind,
		"gnoweb listener",
	)

	fs.StringVar(
		&c.faucetURL,
		"faucet-url",
		defaultWebOptions.faucetURL,
		"The faucet URL will redirect the user when they access `/faucet`.",
	)

	fs.BoolVar(
		&c.json,
		"json",
		defaultWebOptions.json,
		"display log in json format",
	)

	fs.BoolVar(
		&c.html,
		"html",
		defaultWebOptions.html,
		"enable unsafe html",
	)

	fs.BoolVar(
		&c.analytics,
		"with-analytics",
		defaultWebOptions.analytics,
		"nable privacy-first analytics",
	)

	fs.BoolVar(
		&c.noStrict,
		"no-strict",
		defaultWebOptions.noStrict,
		"allow cross-site resource forgery and disable https enforcement",
	)

	fs.BoolVar(
		&c.verbose,
		"v",
		defaultWebOptions.verbose,
		"verbose logging mode",
	)

	fs.DurationVar(
		&c.timeout,
		"timeout",
		defaultWebOptions.timeout,
		"set read/write/idle timeout for server connections",
	)
}

func setupWeb(cfg *webCfg, _ []string, io commands.IO) (func() error, error) {
	// Setup logger
	level := zapcore.InfoLevel
	if cfg.verbose {
		level = zapcore.DebugLevel
	}
	var zapLogger *zap.Logger
	if cfg.json {
		zapLogger = log.NewZapJSONLogger(io.Out(), level)
	} else {
		zapLogger = log.NewZapConsoleLogger(io.Out(), level)
	}
	defer zapLogger.Sync()

	logger := log.ZapLoggerToSlog(zapLogger)

	// Setup app
	appcfg := gnoweb.NewDefaultAppConfig()
	appcfg.ChainID = cfg.chainid
	appcfg.NodeRemote = cfg.remote
	appcfg.RemoteHelp = cfg.remoteHelp
	if appcfg.RemoteHelp == "" {
		appcfg.RemoteHelp = appcfg.NodeRemote
	}
	appcfg.Analytics = cfg.analytics
	appcfg.UnsafeHTML = cfg.html
	appcfg.FaucetURL = cfg.faucetURL
	appcfg.AssetsDir = cfg.assetsDir
	app, err := gnoweb.NewRouter(logger, appcfg)
	if err != nil {
		return nil, fmt.Errorf("unable to start gnoweb app: %w", err)
	}

	// Resolve binding address
	bindaddr, err := net.ResolveTCPAddr("tcp", cfg.bind)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve listener %q: %w", cfg.bind, err)
	}

	logger.Info("Running", "listener", bindaddr.String())

	// Setup security headers
	secureHandler := SecureHeadersMiddleware(app, !cfg.noStrict)

	// Setup server
	server := &http.Server{
		Handler:           secureHandler,
		Addr:              bindaddr.String(),
		ReadTimeout:       cfg.timeout, // Time to read the request
		WriteTimeout:      cfg.timeout, // Time to write the entire response
		IdleTimeout:       cfg.timeout, // Time to keep idle connections open
		ReadHeaderTimeout: time.Minute, // Time to read request headers
	}

	return func() error {
		if err := server.ListenAndServe(); err != nil {
			logger.Error("HTTP server stopped", "error", err)
			return commands.ExitCodeError(1)
		}

		return nil
	}, nil
}

func SecureHeadersMiddleware(next http.Handler, strict bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing by browsers. This ensures that the browser
		// does not interpret files as a different MIME type than declared.
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent the page from being embedded in an iframe. This mitigates
		// clickjacking attacks by ensuring the page cannot be loaded in a frame.
		w.Header().Set("X-Frame-Options", "DENY")

		// Control the amount of referrer information sent in the Referer header.
		// 'no-referrer' ensures that no referrer information is sent, which
		// enhances privacy and prevents leakage of sensitive URLs.
		w.Header().Set("Referrer-Policy", "no-referrer")

		// In `strict` mode, prevent cross-site ressources forgery and enforce https
		if strict {
			// Define a Content Security Policy (CSP) to restrict the sources of
			// scripts, styles, images, and other resources. This helps prevent
			// cross-site scripting (XSS) and other code injection attacks.
			// - 'self' allows resources from the same origin.
			// - 'data:' allows inline images (e.g., base64-encoded images).
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'")

			// Enforce HTTPS by telling browsers to only access the site over HTTPS
			// for a specified duration (1 year in this case). This also applies to
			// subdomains and allows preloading into the browser's HSTS list.
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		next.ServeHTTP(w, r)
	})
}
