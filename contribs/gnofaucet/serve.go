package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
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
	defaultListenAddress = "0.0.0.0:5050"
)

// url & struct for verify captcha
var siteVerifyURL = "https://api.hcaptcha.com/siteverify"

const (
	ipv6Loopback = "::1"
	ipv6ZeroAddr = "0:0:0:0:0:0:0:1"
	ipv4Loopback = "127.0.0.1"
)

var remoteRegex = regexp.MustCompile(`^https?://[a-z\d.-]+(:\d+)?(?:/[a-z\d]+)*$`)

var errInvalidCaptcha = errors.New("unable to verify captcha")

type SiteVerifyResponse struct {
	Success     bool      `json:"success"`
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

	remote        string
	isBehindProxy bool
	logLevel      string
}

func newServeCmd() *commands.Command {
	cfg := &serveCfg{}
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "serve",
			ShortUsage: "<subcommand> [flags]",
			ShortHelp:  "serve <subcommand> [flags]",
			LongHelp:   "Serves the gno.land faucet to users",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newCaptchaCmd(cfg),
		newGithubCmd(cfg),
	)

	return cmd
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
		"10000000ugnot",
		"the static max send amount (native currency)",
	)

	fs.BoolVar(
		&c.isBehindProxy,
		"is-behind-proxy",
		false,
		"use X-Forwarded-For IP for throttling",
	)

	fs.StringVar(
		&c.logLevel,
		"log-level",
		"info",
		"log level (debug, info, warn, error)",
	)
}

// newLogger constructs a JSON structured logger at the configured level.
func (c *serveCfg) newLogger(io commands.IO) (*slog.Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(c.logLevel)); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", c.logLevel, err)
	}

	return log.ZapLoggerToSlog(log.NewZapJSONLogger(io.Out(), level)), nil
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

func serveFaucet(
	ctx context.Context,
	cfg *serveCfg,
	logger *slog.Logger,
	opts ...faucet.Option,
) error {
	// Parse static gas values.
	// It is worth noting that this is temporary,
	// and will be removed once gas estimation is enabled
	// on gno.land
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
	cli, err := tm2Client.NewClient(cfg.remote)
	if err != nil {
		return fmt.Errorf("unable to create TM2 client, %w", err)
	}

	faucetOpts := []faucet.Option{
		faucet.WithLogger(logger),
		faucet.WithConfig(cfg.generateFaucetConfig()),
	}
	faucetOpts = append(faucetOpts, opts...)

	// Create a new faucet with
	// static gas estimation
	f, err := faucet.NewFaucet(
		static.New(gasFee, gasWanted),
		cli,
		faucetOpts...,
	)
	if err != nil {
		return fmt.Errorf("unable to create faucet, %w", err)
	}

	return f.Serve(ctx)
}
