package dial

import (
	"crypto/rand"
	"math/big"
	"net"
	"slices"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/p2p/types"
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

// generateAddr generates a dial address for a fresh peer ID on the given port
func generateAddr(t *testing.T, port uint16) *types.NetAddress {
	t.Helper()

	addr := types.NewNetAddressFromIPPort(net.ParseIP("127.0.0.1"), port)
	addr.ID = types.GenerateNodeKey().ID()

	return addr
}

func TestQueue_Remove(t *testing.T) {
	t.Parallel()

	t.Run("every item for the address is removed", func(t *testing.T) {
		t.Parallel()

		var (
			now    = time.Now()
			target = generateAddr(t, 26656)
			other  = generateAddr(t, 26657)
			q      = NewQueue()
		)

		q.Push(Item{Time: now, Address: target})
		q.Push(Item{Time: now.Add(time.Second), Address: other})
		q.Push(Item{Time: now.Add(2 * time.Second), Address: target})

		q.Remove(target)

		require.Len(t, q.items, 1)
		assert.Equal(t, other, q.items[0].Address)
	})

	t.Run("the remaining items keep their order", func(t *testing.T) {
		t.Parallel()

		var (
			now    = time.Now()
			target = generateAddr(t, 26656)
			first  = generateAddr(t, 26657)
			second = generateAddr(t, 26658)
			q      = NewQueue()
		)

		q.Push(Item{Time: now, Address: first})
		q.Push(Item{Time: now.Add(time.Second), Address: target})
		q.Push(Item{Time: now.Add(2 * time.Second), Address: second})

		q.Remove(target)

		require.Len(t, q.items, 2)
		assert.Equal(t, first, q.Pop().Address)
		assert.Equal(t, second, q.Pop().Address)
	})

	t.Run("the same peer on another address is kept", func(t *testing.T) {
		t.Parallel()

		var (
			now     = time.Now()
			target  = generateAddr(t, 26656)
			sibling = *target
			q       = NewQueue()
		)

		// The same peer ID, on another port
		sibling.Port = 26657

		q.Push(Item{Time: now, Address: target})
		q.Push(Item{Time: now.Add(time.Second), Address: &sibling})

		q.Remove(target)

		require.Len(t, q.items, 1)
		assert.Equal(t, &sibling, q.items[0].Address)
	})

	t.Run("no-op when the address is not queued", func(t *testing.T) {
		t.Parallel()

		var (
			queued = generateAddr(t, 26656)
			q      = NewQueue()
		)

		q.Push(Item{Time: time.Now(), Address: queued})

		q.Remove(generateAddr(t, 26657))

		require.Len(t, q.items, 1)
		assert.Equal(t, queued, q.items[0].Address)
	})
}

func TestQueue_PopDue(t *testing.T) {
	t.Parallel()

	t.Run("empty queue", func(t *testing.T) {
		t.Parallel()

		q := NewQueue()

		assert.Nil(t, q.PopDue(time.Now()))
	})

	t.Run("head not due", func(t *testing.T) {
		t.Parallel()

		var (
			now = time.Now()
			q   = NewQueue()
		)

		q.Push(Item{Time: now.Add(time.Minute), Address: generateAddr(t, 26656)})

		assert.Nil(t, q.PopDue(now))
		assert.Len(t, q.items, 1)
	})

	t.Run("head due", func(t *testing.T) {
		t.Parallel()

		var (
			now   = time.Now()
			due   = generateAddr(t, 26656)
			later = generateAddr(t, 26657)
			q     = NewQueue()
		)

		q.Push(Item{Time: now, Address: due})
		q.Push(Item{Time: now.Add(time.Minute), Address: later})

		item := q.PopDue(now)

		require.NotNil(t, item)
		assert.Equal(t, due, item.Address)
		assert.Len(t, q.items, 1)
	})
}
