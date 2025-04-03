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
	rootCfg        *serveCfg
	ghClientID     string
	maxBalance     int64
	cooldownPeriod time.Duration
}

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
		return fmt.Errorf("github client id is required")
	}

	if cfg.cooldownPeriod <= 0 {
		return fmt.Errorf("cooldown period must be greater than 0")
	}
	clientSecret := os.Getenv("GH_CLIENT_SECRET")
	if clientSecret == "" {
		return fmt.Errorf("github client secret is required")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Username: os.Getenv("REDIS_USER"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})
	err := rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("unable to connect to redis, %w", err)
	}

	// Create cooldown limiter
	cooldownLimiter := NewCooldownLimiter(cfg.cooldownPeriod, rdb)

	return serveFaucet(ctx, cfg.rootCfg, io, getGithubMiddleware(cfg.ghClientID, clientSecret, cooldownLimiter))
}
