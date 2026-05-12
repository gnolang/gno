package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServeFaucet_CleanupShorterThanRateLimit(t *testing.T) {
	t.Parallel()

	cfg := &serveCfg{
		rateLimitInterval:     24 * time.Hour,
		rateLimitCleanTimeout: time.Hour,
	}

	err := serveFaucet(context.Background(), cfg, nil)

	assert.ErrorContains(t, err, "ratelimit-cleanup-timeout must be >= ratelimit-interval")
}
