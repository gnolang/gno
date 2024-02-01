package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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
	defaultListenAddress = "http://127.0.0.1:5050"
)

// url & struct for verify captcha
const siteVerifyURL = "https://www.google.com/recaptcha/api/siteverify"

var remoteRegex = regexp.MustCompile(`^https?://[a-z\d.-]+(:\d+)?(?:/[a-z\d]+)*$`)

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
	sendAmount    string
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
		&c.sendAmount,
		"send",
		"1000000ugnot",
		"the static send amount (native currency)",
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
	cfg.SendAmount = c.sendAmount
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
	_, err = std.ParseCoins(cfg.sendAmount)
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
	logger := log.NewZapJSONLogger(io.Out(), zapcore.DebugLevel)

	// Start throttled faucet.
	st := NewSubnetThrottler()
	if err = st.Start(); err != nil {
		return fmt.Errorf("unable to start throttler service, %w", err)
	}

	// Prepare the middleware
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				host := ""
				if !cfg.isBehindProxy {
					addr := r.RemoteAddr
					host_, _, err := net.SplitHostPort(addr)
					if err != nil {
						return
					}
					host = host_
				} else if xff, found := r.Header["X-Forwarded-For"]; found && len(xff) > 0 {
					host = xff[0]
				}

				// if can't identify the IP, everyone is in the same pool.
				// if host using ipv6 loopback addr, make it ipv4
				if host == "" || host == "::1" || host == "0:0:0:0:0:0:0:1" {
					host = "127.0.0.1"
				}

				ip := net.ParseIP(host)

				if ip == nil {
					io.Println("No IP found")

					http.Error(w, "No IP found", http.StatusUnauthorized)

					return
				}

				allowed, reason := st.Request(ip)
				if !allowed {
					msg := fmt.Sprintf("abuse protection system (%s)", reason)

					io.Println(msg)
					http.Error(w, msg, http.StatusUnauthorized)

					return
				}

				if err = r.ParseForm(); err != nil {
					http.Error(w, "Invalid form", http.StatusBadRequest)

					return
				}

				// only when command line argument 'captcha-secret' has entered > captcha are enabled.
				// verify captcha
				if cfg.captchaSecret != "" {
					passedMsg := r.Form["g-recaptcha-response"]

					if passedMsg == nil {
						http.Error(w, "Invalid captcha request", http.StatusInternalServerError)

						return
					}

					capMsg := strings.TrimSpace(passedMsg[0])

					if err = checkRecaptcha(cfg.captchaSecret, capMsg); err != nil {
						io.Printf("%s recaptcha failed; %v\n", ip, err)

						http.Error(w, "Invalid captcha", http.StatusUnauthorized)

						return
					}
				}

				// Continue with serving the faucet request
				next.ServeHTTP(w, r)
			},
		)
	}

	// Create a new faucet with
	// static gas estimation
	f, err := faucet.NewFaucet(
		static.New(gasFee, gasWanted),
		cli,
		faucet.WithLogger(log.ZapLoggerToSlog(logger)),
		faucet.WithConfig(cfg.generateFaucetConfig()),
		faucet.WithMiddlewares([]faucet.Middleware{middleware}),
	)
	if err != nil {
		return fmt.Errorf("unable to create faucet, %w", err)
	}

	return f.Serve(ctx)
}

// checkRecaptcha checks the recaptcha secret
func checkRecaptcha(secret, response string) error {
	req, err := http.NewRequest(
		http.MethodPost,
		siteVerifyURL,
		nil,
	)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("secret", secret)
	q.Add("response", response)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req) // 200 OK
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body SiteVerifyResponse
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return errors.New("fail, decode response")
	}

	if !body.Success {
		return errors.New("unsuccessful recaptcha verify request")
	}

	return nil
}
