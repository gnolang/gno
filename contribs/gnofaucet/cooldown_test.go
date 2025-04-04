package main

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestCooldownLimiter(t *testing.T) {
	var tenGnots int64 = 10_000_000
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})

	cooldownDuration := time.Second
	limiter := NewCooldownLimiter(cooldownDuration, rdb, 0)
	ctx := context.Background()
	user := "testUser"

	// First check should be allowed
	allowed, err := limiter.CheckCooldown(ctx, user, tenGnots)
	require.NoError(t, err)

	if !allowed {
		t.Errorf("Expected first CheckCooldown to return true, but got false")
	}

	allowed, err = limiter.CheckCooldown(ctx, user, tenGnots)
	require.NoError(t, err)
	// Second check immediately should be denied
	if allowed {
		t.Errorf("Expected second CheckCooldown to return false, but got true")
	}

	require.Eventually(t, func() bool {
		allowed, err := limiter.CheckCooldown(ctx, user, tenGnots)
		return err == nil && !allowed
	}, 2*cooldownDuration, 10*time.Millisecond, "Expected CheckCooldown to return true after cooldown period")
}
