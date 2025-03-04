package ratelimiter

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
)

var (
	addr1 = testutils.TestAddress("alice")
	addr2 = testutils.TestAddress("bob")
)

// TestRateLimiter tests the RateLimiter functionality
func TestRateLimiter(t *testing.T) {
	// Create a rate limiter with max 5 actions and a window size of 100 blocks
	rl := New(5, 100)

	// Set initial block height
	std.TestSkipHeights(150)

	// Test cases using urequire
	urequire.True(t, rl.AllowAction(addr1), "Expected addr1 to be allowed on first action")

	// Allow additional actions for addr1
	for i := 1; i < 5; i++ {
		urequire.True(t, rl.AllowAction(addr1), "Expected addr1 to be allowed within the limit")
	}

	// User1 should be blocked after 5 actions
	urequire.False(t, rl.AllowAction(addr1), "Expected addr1 to be blocked after exceeding the limit")

	// Fast forward blocks beyond the window size
	std.TestSkipHeights(100)

	// User1 should be allowed again after the window size has passed
	urequire.True(t, rl.AllowAction(addr1), "Expected addr1 to be allowed after window size passed")

	// Test actions for another user
	urequire.True(t, rl.AllowAction(addr2), "Expected addr2 to be allowed on first action")

	// Allow additional actions for addr2
	for i := 1; i < 5; i++ {
		urequire.True(t, rl.AllowAction(addr2), "Expected addr2 to be allowed within the limit")
	}

	// User2 should be blocked after 5 actions
	urequire.False(t, rl.AllowAction(addr2), "Expected addr2 to be blocked after exceeding the limit")
}

// TestSetMaxActions tests the SetMaxActions method of RateLimiter
func TestSetMaxActions(t *testing.T) {
	rl := New(2, 100)

	// Set initial block height
	std.TestSkipHeights(150)

	// Test allowing actions within the initial limit
	urequire.True(t, rl.AllowAction(addr1), "Expected addr1 to be allowed on first action")
	urequire.True(t, rl.AllowAction(addr1), "Expected addr1 to be allowed on second action")

	// addr1 should be blocked after exceeding the initial limit
	urequire.False(t, rl.AllowAction(addr1), "Expected addr1 to be blocked after exceeding the limit")

	// Update max actions to 3
	rl.SetMaxActions(3)

	// addr1 should be allowed again with updated limit
	urequire.True(t, rl.AllowAction(addr1), "Expected addr1 to be allowed after updating max actions")
}

// TestSetWindowSize tests the SetWindowSize method of RateLimiter
func TestSetWindowSize(t *testing.T) {
	rl := New(5, 100)

	// Set initial block height
	std.TestSkipHeights(150)

	// Test allowing actions within the initial window size
	urequire.True(t, rl.AllowAction(addr1), "Expected addr1 to be allowed on first action")
	urequire.True(t, rl.AllowAction(addr1), "Expected addr1 to be allowed within the limit")

	// Update window size to a smaller value
	rl.SetWindowSize(2)

	// Fast forward blocks beyond the new window size
	std.TestSkipHeights(2)

	// addr1 should be allowed again with updated window size
	urequire.True(t, rl.AllowAction(addr1), "Expected addr1 to be allowed after updating window size")
}
