package ratelimiter

import (
	"std"

	"gno.land/p/demo/dequeue"
)

// RateLimiter struct to hold the configuration and state of rate limiting
type RateLimiter struct {
	MaxActions     int                              // Maximum actions allowed within the window
	WindowSize     int                              // The size of the window in blocks
	OnchainActions map[std.Address]*dequeue.Dequeue // Maps onchain addresses to action history queue
}

// New creates a new RateLimiter with the given max actions and window size (in blocks)
func New(maxActions, windowSize int) *RateLimiter {
	return &RateLimiter{
		MaxActions:     maxActions,
		WindowSize:     windowSize,
		OnchainActions: make(map[std.Address]*dequeue.Dequeue),
	}
}

// AllowAction checks if an onchain address is allowed to perform an action in the current block
func (rl *RateLimiter) AllowAction(address std.Address) bool {
	currentBlockHeight := std.GetHeight()

	// Get or create the onchain address's action history queue
	actionDeque, exists := rl.OnchainActions[address]
	if !exists {
		actionDeque = dequeue.New()
		rl.OnchainActions[address] = actionDeque
	}

	// Remove actions that are outside the window
	for !actionDeque.Empty() {
		frontBlockHeight, _ := actionDeque.Front()
		if frontBlockHeight.(int64) <= currentBlockHeight-int64(rl.WindowSize) {
			actionDeque.PopFront()
		} else {
			break
		}
	}

	// If the number of valid actions is below the maximum, allow the action
	if actionDeque.Length() < uint64(rl.MaxActions) {
		actionDeque.PushBack(currentBlockHeight)
		return true
	}

	return false
}
