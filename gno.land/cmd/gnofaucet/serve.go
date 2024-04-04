package main

import (
	"context"
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/gnolang/faucet"
	tm2Client "github.com/gnolang/faucet/client/http"
	"github.com/gnolang/faucet/config"
	"github.com/gnolang/faucet/estimate/static"
	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
	"go.uber.org/zap/zapcore"
)

const (
	defaultGasFee        = "1000000ugnot"
	defaultGasWanted     = "100000"
	defaultRemote        = "http://127.0.0.1:26657"
	defaultListenAddress = "127.0.0.1:5050"
)

// url & struct for verify captcha
const siteVerifyURL = "https://www.google.com/recaptcha/api/siteverify"

const (
	ipv6Loopback = "::1"
	ipv6ZeroAddr = "0:0:0:0:0:0:0:1"
	ipv4Loopback = "127.0.0.1"
)

var remoteRegex = regexp.MustCompile(`^https?://[a-z\d.-]+(:\d+)?(?:/[a-z\d]+)*$`)

var errInvalidCaptcha = errors.New("unable to verify captcha")

type SiteVerifyResponse struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

type serveCfg struct {
	listenAddress string
	chainID       string
	mnemonic      string
	maxSendAmount string
	numAccounts   uint64

	remote string

	captchaSecret string
	isBehindProxy bool
}

func newServeCmd() *commands.Command {
	cfg := &serveCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "serve",
			ShortUsage: "serve [flags]",
			LongHelp:   "Serves the gno.land faucet to users",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execServe(ctx, cfg, commands.NewDefaultIO())
		},
	)
}

func (c *serveCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.listenAddress,
		"listen-address",
		defaultListenAddress,
		"the faucet server listen address",
	)

	fs.StringVar(
		&c.remote,
		"remote",
		defaultRemote,
		"remote node URL",
	)

	fs.StringVar(
		&c.mnemonic,
		"mnemonic",
		"",
		"the mnemonic for faucet keys",
	)

	fs.Uint64Var(
		&c.numAccounts,
		"num-accounts",
		1,
		"the number of faucet accounts, based on the mnemonic",
	)

	fs.StringVar(
		&c.chainID,
		"chain-id",
		"",
		"the chain ID associated with the remote Gno chain",
	)

	fs.StringVar(
		&c.maxSendAmount,
		"max-send-amount",
		"1000000ugnot",
		"the static max send amount (native currency)",
	)

	fs.StringVar(
		&c.captchaSecret,
		"captcha-secret",
		"",
		"recaptcha secret key (if empty, captcha are disabled)",
	)

	fs.BoolVar(
		&c.isBehindProxy,
		"is-behind-proxy",
		false,
		"use X-Forwarded-For IP for throttling",
	)
}

// generateFaucetConfig generates the Faucet configuration
// based on the flag data
func (c *serveCfg) generateFaucetConfig() *config.Config {
	// Create the default configuration
	cfg := config.DefaultConfig()

	cfg.ListenAddress = c.listenAddress
	cfg.ChainID = c.chainID
	cfg.Mnemonic = c.mnemonic
	cfg.MaxSendAmount = c.maxSendAmount
	cfg.NumAccounts = c.numAccounts

	return cfg
}

func execServe(ctx context.Context, cfg *serveCfg, io commands.IO) error {
	// Parse static gas values.
	// It is worth noting that this is temporary,
	// and will be removed once gas estimation is enabled
	// on Gno.land
	gasFee := std.MustParseCoin(defaultGasFee)

	gasWanted, err := strconv.ParseInt(defaultGasWanted, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid gas wanted, %w", err)
	}

	// Parse the send amount
	_, err = std.ParseCoins(cfg.maxSendAmount)
	if err != nil {
		return fmt.Errorf("invalid send amount, %w", err)
	}

	// Validate the remote address
	if !remoteRegex.MatchString(cfg.remote) {
		return errors.New("invalid remote address")
	}

	// Create the client (HTTP)
	cli := tm2Client.NewClient(cfg.remote)

	// Set up the logger
	logger := log.ZapLoggerToSlog(
		log.NewZapJSONLogger(
			io.Out(),
			zapcore.DebugLevel,
		),
	)

	// Start throttled faucet.
	st := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)
	st.start(ctx)

	// Prepare the middlewares
	middlewares := []faucet.Middleware{
		getIPMiddleware(cfg.isBehindProxy, st),
		getCaptchaMiddleware(cfg.captchaSecret),
	}

	// Create a new faucet with
	// static gas estimation
	f, err := faucet.NewFaucet(
		static.New(gasFee, gasWanted),
		cli,
		faucet.WithLogger(logger),
		faucet.WithConfig(cfg.generateFaucetConfig()),
		faucet.WithMiddlewares(middlewares),
	)
	if err != nil {
		return fmt.Errorf("unable to create faucet, %w", err)
	}

	return f.Serve(ctx)
}
