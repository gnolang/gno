package main

import (
	"sync"
	"time"
)

// CooldownLimiter is a Limiter using an in-memory map
type CooldownLimiter struct {
	cooldowns    map[string]time.Time
	mu           sync.RWMutex
	cooldownTime time.Duration
}

// NewCooldownLimiter initializes a Cooldown Limiter with a given duration
func NewCooldownLimiter(cooldown time.Duration) *CooldownLimiter {
	return &CooldownLimiter{
		cooldowns:    make(map[string]time.Time),
		cooldownTime: cooldown,
	}
}

// CheckCooldown checks if a key has done some action before the cooldown period has passed
func (rl *CooldownLimiter) CheckCooldown(key string) bool {
	lastClaim := rl.get(key)
	if time.Since(lastClaim) < rl.cooldownTime {
		return false // Deny claim if within cooldown period
	}

	rl.set(key, time.Now())
	return true
}

func (rl *CooldownLimiter) get(key string) time.Time {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.cooldowns[key]
}

func (rl *CooldownLimiter) set(key string, value time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.cooldowns[key] = value
}
