package main

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// CooldownLimiter is a Limiter using an in-memory map
type CooldownLimiter struct {
	cooldownDB   *badger.DB
	cooldownTime time.Duration
}

// NewCooldownLimiter initializes a Cooldown Limiter with a given duration
func NewCooldownLimiter(cooldown time.Duration, dbPath string) *CooldownLimiter {
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		panic(err)
	}

	return &CooldownLimiter{
		cooldownDB:   db,
		cooldownTime: cooldown,
	}
}

// CheckCooldown checks if a key has done some action before the cooldown period has passed
// Returns true if the key is not on cooldown setting
// Returns false if the key is on cooldown
// also marks the key as on cooldown before returning
func (rl *CooldownLimiter) CheckCooldown(key string) (bool, error) {
	isOnCooldown, err := rl.isOnCooldown(key)
	if err != nil {
		return false, fmt.Errorf("unable to check if key is on cooldown, %w", err)
	}
	if isOnCooldown {
		return false, nil // Deny claim if within cooldown period
	}

	return true, rl.markOnCooldown(key)
}

func (rl *CooldownLimiter) isOnCooldown(key string) (bool, error) {
	err := rl.cooldownDB.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		return err
	})

	if err != nil {
		// Real error
		if err != badger.ErrKeyNotFound {
			return false, err
		}
		// errNotFound: key is not on cooldown
		return false, nil
	}

	// Key found: it is on cooldown
	return true, nil
}

func (rl *CooldownLimiter) markOnCooldown(key string) error {
	return rl.cooldownDB.Update(func(txn *badger.Txn) error {
		// Here the value set does not mather as we remain only on the
		// key existence (TTL) to be eligible to claim again
		e := badger.NewEntry([]byte(key), []byte("claimed")).WithTTL(rl.cooldownTime)
		return txn.SetEntry(e)
	})
}
