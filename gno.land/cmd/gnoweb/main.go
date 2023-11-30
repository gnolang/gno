package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	// for static files
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/tm2/pkg/log"
	// for error types
	// "github.com/gnolang/gno/tm2/pkg/sdk"               // for baseapp (info, status)
)

func parseConfigFlags(fs *flag.FlagSet, args []string) (gnoweb.Config, error) {
	cfg := gnoweb.NewDefaultConfig()

	fs.StringVar(&cfg.RemoteAddr, "remote", cfg.RemoteAddr, "remote gnoland node address")
	fs.StringVar(&cfg.CaptchaSite, "captcha-site", cfg.CaptchaSite, "recaptcha site key (if empty, captcha are disabled)")
	fs.StringVar(&cfg.FaucetURL, "faucet-url", cfg.FaucetURL, "faucet server URL")
	fs.StringVar(&cfg.ViewsDir, "views-dir", cfg.ViewsDir, "views directory location") // XXX: replace with goembed
	fs.StringVar(&cfg.HelpChainID, "help-chainid", cfg.HelpChainID, "help page's chainid")
	fs.StringVar(&cfg.HelpRemote, "help-remote", cfg.HelpRemote, "help page's remote addr")
	fs.BoolVar(&cfg.WithAnalytics, "with-analytics", cfg.WithAnalytics, "enable privacy-first analytics")

	return cfg, fs.Parse(args)
}

func main() {
	fs := flag.NewFlagSet("gnoweb", flag.PanicOnError)

	var bindAddress string
	fs.StringVar(&bindAddress, "bind", "127.0.0.1:8888", "server listening address")

	cfg, err := parseConfigFlags(fs, os.Args)
	if err != nil {
		panic("unable to parse flags: " + err.Error())
	}

	logger := log.NewTMLogger(os.Stdout)
	logger.SetLevel(log.LevelDebug)

	logger.Debug("", "Running", "http://"+cfg.BindAddr)
	server := &http.Server{
		Addr:              cfg.BindAddr,
		ReadHeaderTimeout: 60 * time.Second,
		Handler:           gnoweb.MakeApp(logger, cfg).Router,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error("HTTP server stopped", " error:", err)
	}
}
