package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// redisLimiter limits a specific user to one claim per cooldown period
// this limiter keeps track of which keys are on cooldown using a badger database (written to a local file)
type redisLimiter struct {
	redis             *redis.Client
	cooldownTime      time.Duration
	maxlifeTimeAmount *int64
}

// newRedisLimiter initializes a Cooldown Limiter with a given duration
func newRedisLimiter(cooldown time.Duration, redis *redis.Client, maxlifeTimeAmount int64) *redisLimiter {
	limiter := &redisLimiter{
		redis:        redis,
		cooldownTime: cooldown,
	}
	if maxlifeTimeAmount > 0 {
		limiter.maxlifeTimeAmount = &maxlifeTimeAmount
	}

	return limiter
}

// checkCooldown checks if a key can make a claim or if it is still within the cooldown period
// also checks that the user will not exceed the max lifetime allowed amount
// Returns true if the key is not on cooldown, and marks the key as on cooldown
// Returns false if the key is on cooldown or if an error occurs
func (rl *redisLimiter) checkCooldown(ctx context.Context, key string, amountClaimed int64) (bool, error) {
	claimData, err := rl.getClaimsData(ctx, key)
	if err != nil {
		return false, fmt.Errorf("unable to check if key is on cooldown, %w", err)
	}
	// Deny claim if within cooldown period
	if claimData.LastClaimed.Add(rl.cooldownTime).After(time.Now()) {
		return false, nil
	}
	// check that user will not exceed max lifetime allowed amount
	if rl.maxlifeTimeAmount != nil && claimData.TotalClaimed+amountClaimed > *rl.maxlifeTimeAmount {
		return false, nil
	}

	return true, rl.declareClaimedValue(ctx, key, amountClaimed, claimData)
}

func (rl *redisLimiter) getClaimsData(ctx context.Context, key string) (*claimData, error) {
	storedData, err := rl.redis.Get(ctx, key).Result()
	if err != nil {
		// Here we return an empty claimData because is the first time the user is making a claim
		// the total amount claimed is 0 and the lastClaimed is the default time value
		if errors.Is(err, redis.Nil) {
			return &claimData{}, nil
		}
		// Any other unexpected error is returned.
		return nil, err
	}

	claimData := &claimData{}
	err = json.Unmarshal([]byte(storedData), claimData)
	return claimData, err
}

func (rl *redisLimiter) declareClaimedValue(ctx context.Context, key string, amountClaimed int64, currentData *claimData) error {
	currentData.LastClaimed = time.Now()
	currentData.TotalClaimed += amountClaimed

	data, err := json.Marshal(currentData)
	if err != nil {
		return fmt.Errorf("unable to marshal claim data, %w", err)
	}

	return rl.redis.Set(ctx, key, data, 0).Err()
}

type claimData struct {
	LastClaimed  time.Time
	TotalClaimed int64
}
