package github

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRewarder(t *testing.T) (*RedisRewarder, *redis.Client) {
	t.Helper()

	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})

	cfg := &RewarderCfg{
		MaxReward:    1000,
		PRFactor:     10.0,
		ReviewFactor: 5.0,
		IssueFactor:  3.0,
		CommitFactor: 1.0,
	}

	rewarder := &RedisRewarder{
		redisClient: rdb,
		cfg:         cfg,
	}

	t.Cleanup(redisServer.Close)

	return rewarder, rdb
}

func TestGetReward_SumNeverExceedsMaxReward(t *testing.T) {
	rewarder, rdb := setupTestRewarder(t)

	ctx := context.Background()
	user := "testuser"

	rdb.Set(ctx, issueCountKey(user), 10, 0)   // 10 * 3 = 30
	rdb.Set(ctx, prCountKey(user), 5, 0)       // 5 * 10 = 50
	rdb.Set(ctx, prReviewCountKey(user), 8, 0) // 8 * 5 = 40
	rdb.Set(ctx, commitCountKey(user), 20, 0)  // 20 * 1 = 20
	// Total: 30 + 50 + 40 + 20 = 140

	reward, err := rewarder.GetReward(ctx, user)
	require.NoError(t, err)
	assert.Equal(t, 140, reward)
	err = rewarder.Apply(ctx, user, reward)
	require.NoError(t, err)

	rdb.Set(ctx, issueCountKey(user), 100, 0)   // 100 * 3 = 300
	rdb.Set(ctx, prCountKey(user), 50, 0)       // 50 * 10 = 500
	rdb.Set(ctx, prReviewCountKey(user), 80, 0) // 80 * 5 = 400
	rdb.Set(ctx, commitCountKey(user), 200, 0)  // 200 * 1 = 200
	// Total: 300 + 500 + 400 + 200 = 1400, but MaxReward is 1000

	reward, err = rewarder.GetReward(ctx, user)
	require.NoError(t, err)
	assert.Equal(t, 860, reward) // 1000 - 140 = 860
	err = rewarder.Apply(ctx, user, reward)
	require.NoError(t, err)

	reward, err = rewarder.GetReward(ctx, user)
	require.NoError(t, err)
	assert.Equal(t, 0, reward)

	totalRewarded, err := rdb.Get(ctx, userRewardedKey(user)).Int()
	require.NoError(t, err)
	assert.Equal(t, 1000, totalRewarded)
}

func TestGetReward_EdgeCases(t *testing.T) {
	rewarder, rdb := setupTestRewarder(t)

	ctx := context.Background()
	user := "testuser"

	t.Run("zero Counts", func(t *testing.T) {
		rdb.Set(ctx, issueCountKey(user), 0, 0)
		rdb.Set(ctx, prCountKey(user), 0, 0)
		rdb.Set(ctx, prReviewCountKey(user), 0, 0)
		rdb.Set(ctx, commitCountKey(user), 0, 0)

		reward, err := rewarder.GetReward(ctx, user)
		require.NoError(t, err)
		assert.Equal(t, 0, reward)
	})

	t.Run("missing keys", func(t *testing.T) {
		rdb.Del(ctx, issueCountKey(user), prCountKey(user), prReviewCountKey(user), commitCountKey(user), userRewardedKey(user))

		reward, err := rewarder.GetReward(ctx, user)
		require.NoError(t, err)
		assert.Equal(t, 0, reward)
	})

	t.Run("large numbers", func(t *testing.T) {
		rdb.Del(ctx, issueCountKey(user), prCountKey(user), prReviewCountKey(user), commitCountKey(user), userRewardedKey(user))

		rdb.Set(ctx, issueCountKey(user), 1000000, 0) // 1000000 * 3 = 3000000

		reward, err := rewarder.GetReward(ctx, user)
		require.NoError(t, err)
		assert.Equal(t, 1000, reward)
	})
}

func TestGetReward_MultipleUsers(t *testing.T) {
	rewarder, rdb := setupTestRewarder(t)

	ctx := context.Background()
	user1 := "user1"
	user2 := "user2"

	rdb.Set(ctx, issueCountKey(user1), 5, 0)  // 5 * 3 = 15
	rdb.Set(ctx, prCountKey(user1), 2, 0)     // 2 * 10 = 20
	rdb.Set(ctx, issueCountKey(user2), 10, 0) // 10 * 3 = 30
	rdb.Set(ctx, prCountKey(user2), 5, 0)     // 5 * 10 = 50

	reward1, err := rewarder.GetReward(ctx, user1)
	require.NoError(t, err)
	assert.Equal(t, 35, reward1) // 15 + 20
	err = rewarder.Apply(ctx, user1, reward1)
	require.NoError(t, err)

	reward2, err := rewarder.GetReward(ctx, user2)
	require.NoError(t, err)
	assert.Equal(t, 80, reward2) // 30 + 50
	err = rewarder.Apply(ctx, user2, reward2)
	require.NoError(t, err)

	total1, err := rdb.Get(ctx, userRewardedKey(user1)).Int()
	require.NoError(t, err)
	assert.Equal(t, 35, total1)

	total2, err := rdb.Get(ctx, userRewardedKey(user2)).Int()
	require.NoError(t, err)
	assert.Equal(t, 80, total2)
}

func TestGetReward_Rounding(t *testing.T) {
	rewarder, rdb := setupTestRewarder(t)

	ctx := context.Background()
	user := "testuser"

	cfg := &RewarderCfg{
		MaxReward:    1000,
		PRFactor:     1.5,
		ReviewFactor: 2.7,
		IssueFactor:  0.8,
		CommitFactor: 0.3,
	}

	rewarder.cfg = cfg

	rdb.Set(ctx, issueCountKey(user), 3, 0)    // 3 * 0.8 = 2.4 -> 2
	rdb.Set(ctx, prCountKey(user), 4, 0)       // 4 * 1.5 = 6.0 -> 6
	rdb.Set(ctx, prReviewCountKey(user), 5, 0) // 5 * 2.7 = 13.5 -> 14
	rdb.Set(ctx, commitCountKey(user), 7, 0)   // 7 * 0.3 = 2.1 -> 2
	// Expected total: 2 + 6 + 14 + 2 = 24

	reward, err := rewarder.GetReward(ctx, user)
	require.NoError(t, err)
	assert.Equal(t, 24, reward)
}
