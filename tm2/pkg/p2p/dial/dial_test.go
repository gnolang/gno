package dial

import (
	"crypto/rand"
	"math/big"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateRandomTimes generates random time intervals
func generateRandomTimes(t *testing.T, count int) []time.Time {
	t.Helper()

	const timeRange = 94608000 // 3 years

	var (
		maxRange = big.NewInt(time.Now().Unix() - timeRange)
		times    = make([]time.Time, 0, count)
	)

	for range count {
		n, err := rand.Int(rand.Reader, maxRange)
		require.NoError(t, err)

		randTime := time.Unix(n.Int64()+timeRange, 0)

		times = append(times, randTime)
	}

	return times
}

func TestQueue_Push(t *testing.T) {
	t.Parallel()

	var (
		timestamps = generateRandomTimes(t, 10)
		q          = NewQueue()
	)

	// Add the dial items
	for _, timestamp := range timestamps {
		q.Push(Item{
			Time: timestamp,
		})
	}

	assert.Len(t, q.items, len(timestamps))
}

func TestQueue_Peek(t *testing.T) {
	t.Parallel()

	t.Run("empty queue", func(t *testing.T) {
		t.Parallel()

		q := NewQueue()

		assert.Nil(t, q.Peek())
	})

	t.Run("existing item", func(t *testing.T) {
		t.Parallel()

		var (
			timestamps = generateRandomTimes(t, 100)
			q          = NewQueue()
		)

		// Add the dial items
		for _, timestamp := range timestamps {
			q.Push(Item{
				Time: timestamp,
			})
		}

		// Sort the initial list to find the best timestamp
		slices.SortFunc(timestamps, func(a, b time.Time) int {
			if a.Before(b) {
				return -1
			}

			if a.After(b) {
				return 1
			}

			return 0
		})

		assert.Equal(t, q.Peek().Time.Unix(), timestamps[0].Unix())
	})
}

func TestQueue_Pop(t *testing.T) {
	t.Parallel()

	t.Run("empty queue", func(t *testing.T) {
		t.Parallel()

		q := NewQueue()

		assert.Nil(t, q.Pop())
	})

	t.Run("existing item", func(t *testing.T) {
		t.Parallel()

		var (
			timestamps = generateRandomTimes(t, 100)
			q          = NewQueue()
		)

		// Add the dial items
		for _, timestamp := range timestamps {
			q.Push(Item{
				Time: timestamp,
			})
		}

		assert.Len(t, q.items, len(timestamps))

		// Sort the initial list to find the best timestamp
		slices.SortFunc(timestamps, func(a, b time.Time) int {
			if a.Before(b) {
				return -1
			}

			if a.After(b) {
				return 1
			}

			return 0
		})

		for index, timestamp := range timestamps {
			item := q.Pop()

			require.Len(t, q.items, len(timestamps)-1-index)

			assert.Equal(t, item.Time.Unix(), timestamp.Unix())
		}
	})
}
