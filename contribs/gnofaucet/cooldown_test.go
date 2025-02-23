package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCooldownLimiter(t *testing.T) {
	cooldownDuration := time.Second
	limiter := NewCooldownLimiter(cooldownDuration)
	user := "testUser"

	// First check should be allowed
	if !limiter.CheckCooldown(user) {
		t.Errorf("Expected first CheckCooldown to return true, but got false")
	}

	// Second check immediately should be denied
	if limiter.CheckCooldown(user) {
		t.Errorf("Expected second CheckCooldown to return false, but got true")
	}

	require.Eventually(t, func() bool {
		return limiter.CheckCooldown(user)
	}, 2*cooldownDuration, 10*time.Millisecond, "Expected CheckCooldown to return true after cooldown period")
}
