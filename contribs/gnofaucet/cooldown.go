package main

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/pkg/errors"
)

// CooldownLimiter is a limits a specific user to one claim per cooldown period
// this limiter keeps track of which keys are on cooldown using a badger database (written to a local file)
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

// CheckCooldown checks if a key can make a claim or if it is still within the cooldown period
// Returns true if the key is not on cooldown, and marks the key as on cooldown
// Returns false if the key is on cooldown or if an error occurs
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
		// Since we use badger's TTL feature to manage cooldown periods,
		// an ErrKeyNotFound simply indicates that the key is not on cooldown.
		if errors.Is(err, badger.ErrKeyNotFound) {
			return false, nil
		}

		// Any other unexpected error is returned.
		return false, err
	}

	// Key found: it is on cooldown
	return true, nil
}

func (rl *CooldownLimiter) markOnCooldown(key string) error {
	return rl.cooldownDB.Update(func(txn *badger.Txn) error {
		// The value set here does not matter, as we only rely on
		// badger's TTL feature to check if a key is still on cooldown
		e := badger.NewEntry([]byte(key), []byte("claimed")).WithTTL(rl.cooldownTime)
		return txn.SetEntry(e)
	})
}
