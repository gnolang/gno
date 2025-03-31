package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// CooldownLimiter limits a specific user to one claim per cooldown period
// this limiter keeps track of which keys are on cooldown using a badger database (written to a local file)
type CooldownLimiter struct {
	redis        *redis.Client
	cooldownTime time.Duration
}

// NewCooldownLimiter initializes a Cooldown Limiter with a given duration
func NewCooldownLimiter(cooldown time.Duration, redis *redis.Client) *CooldownLimiter {
	return &CooldownLimiter{
		redis:        redis,
		cooldownTime: cooldown,
	}
}

// CheckCooldown checks if a key can make a claim or if it is still within the cooldown period
// Returns true if the key is not on cooldown, and marks the key as on cooldown
// Returns false if the key is on cooldown or if an error occurs
func (rl *CooldownLimiter) CheckCooldown(ctx context.Context, key string) (bool, error) {
	isOnCooldown, err := rl.isOnCooldown(ctx, key)
	if err != nil {
		return false, fmt.Errorf("unable to check if key is on cooldown, %w", err)
	}
	if isOnCooldown {
		return false, nil // Deny claim if within cooldown period
	}

	return true, rl.markOnCooldown(ctx, key)
}

func (rl *CooldownLimiter) isOnCooldown(ctx context.Context, key string) (bool, error) {
	_, err := rl.redis.Get(ctx, key).Result()
	if err != nil {
		// Since we use redis's TTL feature to manage cooldown periods,
		// an error redis.Nil simply indicates that the key is not on cooldown.
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		// Any other unexpected error is returned.
		return false, err
	}

	// Key found: it is on cooldown
	return true, nil
}

func (rl *CooldownLimiter) markOnCooldown(ctx context.Context, key string) error {
	// The value set here does not matter, as we only rely on
	// redis's TTL feature to check if a key is still on cooldown
	return rl.redis.Set(ctx, key, "claimed", rl.cooldownTime).Err()

}
