package main

import (
	"sync"
	"time"
)

// CooldownLimiter is a Limiter using an in-memory map
type CooldownLimiter struct {
	cooldowns    map[string]time.Time
	mu           sync.Mutex
	cooldownTime time.Duration
}

// NewCooldownLimiter initializes a Cooldown Limiter with a given duration
func NewCooldownLimiter(cooldown time.Duration) *CooldownLimiter {
	return &CooldownLimiter{
		cooldowns:    make(map[string]time.Time),
		cooldownTime: cooldown,
	}
}

// CheckCooldown checks if a user is eligible for a reward claim
func (rl *CooldownLimiter) CheckCooldown(ghLogin string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if lastClaim, found := rl.cooldowns[ghLogin]; found {
		if time.Since(lastClaim) < rl.cooldownTime {
			return false // Deny claim if within cooldown period
		}
	}

	rl.cooldowns[ghLogin] = time.Now()
	return true
}
