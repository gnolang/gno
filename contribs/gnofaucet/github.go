package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/gnolang/faucet"
	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/google/go-github/v74/github"
	"github.com/jferrl/go-githubauth"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap/zapcore"
	"golang.org/x/oauth2"

	igh "github.com/gnolang/gno/contribs/gnofaucet/github"
)

type githubCfg struct {
	rootCfg           *serveCfg
	maxClaimableLimit int64
	cooldownPeriod    time.Duration
}

var (
	errGithubClientIDMissing     = fmt.Errorf("GitHub client ID is required")
	errGithubAppIDMissing        = fmt.Errorf("GitHub application ID is required")
	errGithubAppPrivKeyMissing   = fmt.Errorf("GitHub application Private Key is required")
	errGithubClientSecretMissing = fmt.Errorf("GitHub client secret is required")
	errCooldownPeriodInvalid     = fmt.Errorf("cooldown period must be greater than 0")
)

const (
	envGithubClientID             = "GH_CLIENT_ID"
	envGithubClientSecret         = "GH_CLIENT_SECRET"
	envGithubAppID                = "GH_APP_ID"
	envGithubClientPrivateKeyPath = "GH_APP_PRIVATE_KEY_PATH"

	envRedisAddr     = "REDIS_ADDR"
	envRedisUser     = "REDIS_USER"
	envRedisPassword = "REDIS_PASSWORD"
	envFetcherRepos  = "GH_FETCHER_REPOS"

	envFetcherMaxReward    = "GH_FETCHER_MAX_REWARD"
	envFetcherPRFactor     = "GH_FETCHER_PR_FACTOR"
	envFetcherReviewFactor = "GH_FETCHER_REVIEW_FACTOR"
	envFetcherIssueFactor  = "GH_FETCHER_ISSUE_FACTOR"
	envFetcherCommitFactor = "GH_FETCHER_COMMIT_FACTOR"
)

func (c *githubCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.DurationVar(
		&c.cooldownPeriod,
		"cooldown-period",
		24*time.Hour,
		"minimum required time between consecutive faucet claims by the same user",
	)

	fs.Int64Var(
		&c.maxClaimableLimit,
		"max-claimable-limit",
		0,
		"maximum number of tokens a single user can claim over their lifetime. Zero means no limit",
	)
}

type ghFetcherCfg struct {
	rootCfg       *serveCfg
	fetchInterval time.Duration
}

func (c *ghFetcherCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.DurationVar(
		&c.fetchInterval,
		"fetch-interval",
		20*time.Second,
		"polling time to query GitHub API for new events",
	)
}

func newGithubCmd(rootCfg *serveCfg) *commands.Command {
	cfg := &githubCfg{
		rootCfg: rootCfg,
	}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "github",
			ShortUsage: "github [flags]",
			LongHelp:   "applies github middleware to the gno.land faucet",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execGithub(ctx, cfg, commands.NewDefaultIO())
		},
	)

	fcfg := &ghFetcherCfg{
		rootCfg: rootCfg,
	}

	cmd.AddSubCommands(commands.NewCommand(
		commands.Metadata{
			Name:       "fetcher",
			ShortUsage: "fetcher [flags]",
			LongHelp:   "start fetching metadata from specified github repositories",
		},
		fcfg,
		func(ctx context.Context, args []string) error {
			return execGHFetcher(ctx, fcfg, commands.NewDefaultIO())
		},
	))

	return cmd
}

func execGithub(ctx context.Context, cfg *githubCfg, io commands.IO) error {
	clientID := os.Getenv(envGithubClientID)
	if clientID == "" {
		return errGithubClientIDMissing
	}

	if cfg.cooldownPeriod <= 0 {
		return errCooldownPeriodInvalid
	}
	clientSecret := os.Getenv(envGithubClientSecret)
	if clientSecret == "" {
		return errGithubClientSecretMissing
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv(envRedisAddr),
		Username: os.Getenv(envRedisUser),
		Password: os.Getenv(envRedisPassword),
	})
	err := rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("unable to connect to redis, %w", err)
	}

	// Create cooldown limiter
	cooldownLimiter := newRedisLimiter(cfg.cooldownPeriod, rdb, cfg.maxClaimableLimit)

	// Start the IP throttler
	st := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)
	st.start(ctx)

	rewarderCfg, err := parseRewarderConfig()
	if err != nil {
		return fmt.Errorf("failed to parse rewarder config: %w", err)
	}

	rr := igh.NewRedisRewarder(rdb, rewarderCfg)

	logger := log.ZapLoggerToSlog(
		log.NewZapJSONLogger(
			io.Out(),
			zapcore.DebugLevel,
		),
	)

	// Prepare the middlewares
	httpMiddlewares := []func(http.Handler) http.Handler{
		ipMiddleware(cfg.rootCfg.isBehindProxy, st),
		gitHubUsernameMiddleware(clientID, clientSecret, defaultGHExchange, logger, rdb),
	}

	rpcMiddlewares := getMiddlewares(rr, cooldownLimiter)

	return serveFaucet(
		ctx,
		cfg.rootCfg,
		io,
		faucet.WithHTTPMiddlewares(httpMiddlewares),
		faucet.WithMiddlewares(rpcMiddlewares),
	)
}

func execGHFetcher(ctx context.Context, cfg *ghFetcherCfg, io commands.IO) error {
	appIDStr := os.Getenv(envGithubAppID)
	if appIDStr == "" {
		return errGithubAppIDMissing
	}

	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		return err
	}

	pkPath := os.Getenv(envGithubClientPrivateKeyPath)
	if pkPath == "" {
		return errGithubAppPrivKeyMissing
	}

	privKey, err := os.ReadFile(pkPath)
	if err != nil {
		return err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv(envRedisAddr),
		Username: os.Getenv(envRedisUser),
		Password: os.Getenv(envRedisPassword),
	})

	err = rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("unable to connect to redis, %w", err)
	}

	logger := log.ZapLoggerToSlog(
		log.NewZapJSONLogger(
			io.Out(),
			zapcore.DebugLevel,
		),
	)

	appTokenSource, err := githubauth.NewApplicationTokenSource(appID, privKey)
	if err != nil {
		return err
	}

	installationTokenSource := githubauth.NewInstallationTokenSource(78199441, appTokenSource)
	httpClient := oauth2.NewClient(context.Background(), installationTokenSource)
	githubClient := github.NewClient(httpClient)
	githubGraphql := graphql.NewClient("https://api.github.com/graphql", httpClient)
	ghImpl := igh.NewGithubClientImpl(githubClient, githubGraphql)

	fetcher := igh.NewGHFetcher(ghImpl, rdb, parseRepos(), logger, cfg.fetchInterval)

	return fetcher.Fetch(ctx)
}

func parseRepos() map[string][]string {
	reposRaw := os.Getenv(envFetcherRepos)
	if reposRaw == "" {
		panic("set at least 1 github repository to use the fetcher using " + envFetcherRepos + " env var")
	}

	out := make(map[string][]string)
	for _, fullRepo := range strings.Split(reposRaw, " ") {
		orgAndName := strings.Split(fullRepo, "/")
		if len(orgAndName) != 2 {
			panic("repository format must be OWNER/REPONAME")
		}
		owner := orgAndName[0]
		repo := orgAndName[1]

		out[owner] = append(out[owner], repo)
	}

	return out
}

func parseRewarderConfig() (*igh.RewarderCfg, error) {
	cfg := &igh.RewarderCfg{}

	parsers := []struct {
		envVar string
		field  *float64
		name   string
	}{
		{envFetcherPRFactor, &cfg.PRFactor, "PR factor"},
		{envFetcherReviewFactor, &cfg.ReviewFactor, "review factor"},
		{envFetcherIssueFactor, &cfg.IssueFactor, "issue factor"},
		{envFetcherCommitFactor, &cfg.CommitFactor, "commit factor"},
	}

	for _, p := range parsers {
		if val := os.Getenv(p.envVar); val != "" {
			if parsed, err := strconv.ParseFloat(val, 64); err != nil {
				return nil, fmt.Errorf("invalid %s value: %s", p.name, val)
			} else {
				*p.field = parsed
			}
		}
	}

	if maxRewardStr := os.Getenv(envFetcherMaxReward); maxRewardStr != "" {
		if maxReward, err := strconv.Atoi(maxRewardStr); err != nil {
			return nil, fmt.Errorf("invalid max reward value: %s", maxRewardStr)
		} else {
			cfg.MaxReward = maxReward
		}
	}

	return cfg, nil
}
