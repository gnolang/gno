package ratelimiter

import (
	"std"

	"gno.land/p/demo/dequeue"
)

// RateLimiter struct to hold the configuration and state of rate limiting
type RateLimiter struct {
	maxActions     int                              // Maximum actions allowed within the window
	windowSize     int                              // The size of the window in blocks
	onchainActions map[std.Address]*dequeue.Dequeue // Maps onchain addresses to action history queue
}

// New creates a new RateLimiter with the given max actions and window size (in blocks)
func New(maxActions, windowSize int) *RateLimiter {
	return &RateLimiter{
		maxActions:     maxActions,
		windowSize:     windowSize,
		onchainActions: make(map[std.Address]*dequeue.Dequeue),
	}
}

// AllowAction checks if an onchain address is allowed to perform an action in the current block
func (rl *RateLimiter) AllowAction(address std.Address) bool {
	currentBlockHeight := std.GetHeight()

	// Get or create the onchain address's action history queue
	actionDeque, exists := rl.onchainActions[address]
	if !exists {
		actionDeque = dequeue.New()
		rl.onchainActions[address] = actionDeque
	}

	// Remove actions that are outside the window
	for !actionDeque.Empty() {
		earliestBlockHeight, _ := actionDeque.Front()
		if earliestBlockHeight.(int64) <= currentBlockHeight-int64(rl.windowSize) {
			actionDeque.PopFront()
		} else {
			break
		}
	}

	// If the number of valid actions is below the maximum, allow the action
	if actionDeque.Length() < uint64(rl.maxActions) {
		actionDeque.PushBack(currentBlockHeight)
		return true
	}

	return false
}

// SetMaxActions updates the maximum number of actions allowed within the time window.
func (rl *RateLimiter) SetMaxActions(maxActions int) {
	rl.maxActions = maxActions
}

// SetWindowSize updates the size of the time window in blocks.
func (rl *RateLimiter) SetWindowSize(windowSize int) {
	rl.windowSize = windowSize
}
