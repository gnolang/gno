package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/redis/go-redis/v9"
)

type githubCfg struct {
	rootCfg           *serveCfg
	ghClientID        string
	maxClaimableLimit int64
	cooldownPeriod    time.Duration
}

var (
	errGithubClientIDMissing     = fmt.Errorf("GitHub client ID is required")
	errGithubClientSecretMissing = fmt.Errorf("GitHub client secret is required")
	errCooldownPeriodInvalid     = fmt.Errorf("cooldown period must be greater than 0")
)

const (
	envGithubClientSecret = "GH_CLIENT_SECRET"
	envRedisAddr          = "REDIS_ADDR"
	envRedisUser          = "REDIS_USER"
	envRedisPassword      = "REDIS_PASSWORD"
)

func (c *githubCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.ghClientID,
		"github-client-id",
		"",
		"github client id for oauth authentication",
	)

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

func newGithubCmd(rootCfg *serveCfg) *commands.Command {
	cfg := &githubCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
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
}

func execGithub(ctx context.Context, cfg *githubCfg, io commands.IO) error {
	if cfg.ghClientID == "" {
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
	cooldownLimiter := NewCooldownLimiter(cfg.cooldownPeriod, rdb, cfg.maxClaimableLimit)

	return serveFaucet(ctx, cfg.rootCfg, io, getGithubMiddleware(cfg.ghClientID, clientSecret, cooldownLimiter))
}
