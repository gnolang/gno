package github

import (
	"context"
	"fmt"
	"math"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type Rewarder interface {
	GetReward(ctx context.Context, user string) (int, error)
	Apply(ctx context.Context, user string, amount int) error
}

var _ Rewarder = &RedisRewarder{}

type RewarderCfg struct {
	MaxReward int

	PRFactor     float64 // Merged PRs
	ReviewFactor float64 // Approved, Changes-Requested PR Reviews
	IssueFactor  float64 // Opened issues
	CommitFactor float64 // Commits merged into main branch
}

type RedisRewarder struct {
	redisClient *redis.Client
	cfg         *RewarderCfg
}

func NewRedisRewarder(cli *redis.Client, cfg *RewarderCfg) *RedisRewarder {
	return &RedisRewarder{
		redisClient: cli,
		cfg:         cfg,
	}
}

// Reward implements Rewarder.
func (r *RedisRewarder) GetReward(ctx context.Context, user string) (int, error) {
	keys := map[string]float64{
		issueCountKey(user):    r.cfg.IssueFactor,
		prCountKey(user):       r.cfg.PRFactor,
		prReviewCountKey(user): r.cfg.ReviewFactor,
		commitCountKey(user):   r.cfg.CommitFactor,
	}

	var sum float64
	for k, f := range keys {
		c, err := r.getCount(ctx, k)
		if err != nil {
			return 0, err
		}

		sum += float64(c) * f
	}

	total := int(math.Round(sum))

	previouslyRewarded, err := r.getCount(ctx, userRewardedKey(user))
	if err != nil {
		return 0, err
	}

	if total+previouslyRewarded >= r.cfg.MaxReward {
		// we cap the amount to give to maxReward
		total = r.cfg.MaxReward - previouslyRewarded
	} else {
		// if we didn't reach the max amount, we just remove previously rewarded tokens
		total = total - previouslyRewarded
	}

	return total, nil
}

func (r *RedisRewarder) Apply(ctx context.Context, user string, amount int) error {
	previouslyRewarded, err := r.getCount(ctx, userRewardedKey(user))
	if err != nil {
		return err
	}

	return r.redisClient.Set(ctx, userRewardedKey(user), amount+previouslyRewarded, 0).Err()
}

func (r *RedisRewarder) getCount(ctx context.Context, key string) (int, error) {
	out, err := r.redisClient.Get(ctx, key).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, fmt.Errorf("error obtaining redis key: %w", err)
	}

	return out, nil
}

func userRewardedKey(user string) string {
	return fmt.Sprintf("reward:%s", user)
}
