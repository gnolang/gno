package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCooldownLimiter(t *testing.T) {
	testDir := os.TempDir()
	cooldownDuration := time.Second
	limiter, err := NewCooldownLimiter(cooldownDuration, testDir+"/db")
	require.NoError(t, err)
	user := "testUser"

	// First check should be allowed
	allowed, err := limiter.CheckCooldown(user)
	require.NoError(t, err)

	if !allowed {
		t.Errorf("Expected first CheckCooldown to return true, but got false")
	}

	allowed, err = limiter.CheckCooldown(user)
	require.NoError(t, err)
	// Second check immediately should be denied
	if allowed {
		t.Errorf("Expected second CheckCooldown to return false, but got true")
	}

	require.Eventually(t, func() bool {
		allowed, err := limiter.CheckCooldown(user)
		return err == nil && !allowed
	}, 2*cooldownDuration, 10*time.Millisecond, "Expected CheckCooldown to return true after cooldown period")
}
