package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/gnolang/faucet"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type captchaCfg struct {
	rootCfg       *serveCfg
	captchaSecret string
}

var errCaptchaMissing = fmt.Errorf("captcha secret is required")

func (c *captchaCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.captchaSecret,
		"captcha-secret",
		"",
		"recaptcha secret key (if empty, captcha are disabled)",
	)
}

func newCaptchaCmd(rootCfg *serveCfg) *commands.Command {
	cfg := &captchaCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "captcha",
			ShortUsage: "captcha [flags]",
			LongHelp:   "applies captcha middleware to the gno.land faucet",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execCaptcha(ctx, cfg, commands.NewDefaultIO())
		},
	)
}

func execCaptcha(ctx context.Context, cfg *captchaCfg, io commands.IO) error {
	if cfg.captchaSecret == "" {
		return errCaptchaMissing
	}

	// Start the IP throttler
	st := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)
	st.start(ctx)

	// Prepare the middlewares
	httpMiddlewares := []func(http.Handler) http.Handler{
		ipMiddleware(cfg.rootCfg.isBehindProxy, st),
	}

	rpcMiddlewares := []faucet.Middleware{
		captchaMiddleware(cfg.captchaSecret),
	}

	return serveFaucet(
		ctx,
		cfg.rootCfg,
		io,
		faucet.WithHTTPMiddlewares(httpMiddlewares),
		faucet.WithMiddlewares(rpcMiddlewares),
	)
}
