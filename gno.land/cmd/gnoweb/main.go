package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	// for static files
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/log"
	"go.uber.org/zap/zapcore"
	// for error types
	// "github.com/gnolang/gno/tm2/pkg/sdk"               // for baseapp (info, status)
)

func main() {
	err := runMain(os.Args[1:])
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

func runMain(args []string) error {
	var (
		fs          = flag.NewFlagSet("gnoweb", flag.ContinueOnError)
		cfg         = gnoweb.NewDefaultConfig()
		bindAddress string
	)
	fs.StringVar(&cfg.RemoteAddr, "remote", cfg.RemoteAddr, "remote gnoland node address")
	fs.StringVar(&cfg.CaptchaSite, "captcha-site", cfg.CaptchaSite, "recaptcha site key (if empty, captcha are disabled)")
	fs.StringVar(&cfg.FaucetURL, "faucet-url", cfg.FaucetURL, "faucet server URL")
	fs.StringVar(&cfg.ViewsDir, "views-dir", cfg.ViewsDir, "views directory location") // XXX: replace with goembed
	fs.StringVar(&cfg.HelpChainID, "help-chainid", cfg.HelpChainID, "help page's chainid")
	fs.StringVar(&cfg.HelpRemote, "help-remote", cfg.HelpRemote, "help page's remote addr")
	fs.BoolVar(&cfg.WithAnalytics, "with-analytics", cfg.WithAnalytics, "enable privacy-first analytics")
	fs.StringVar(&bindAddress, "bind", "127.0.0.1:8888", "server listening address")

	if err := fs.Parse(args); err != nil {
		return err
	}

	zapLogger := log.NewZapConsoleLogger(os.Stdout, zapcore.DebugLevel)
	logger := log.ZapLoggerToSlog(zapLogger)

	logger.Info("Running", "listener", "http://"+bindAddress)
	server := &http.Server{
		Addr:              bindAddress,
		ReadHeaderTimeout: 60 * time.Second,
		Handler:           gnoweb.MakeApp(logger, cfg).Router,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error("HTTP server stopped", " error:", err)
	}

	return zapLogger.Sync()
}
