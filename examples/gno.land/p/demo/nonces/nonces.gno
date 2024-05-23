package nonce

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type NonceManager struct {
	nonces *avl.Tree
}

// New creates a new instance of NonceManager
func New() *NonceManager {
	return &NonceManager{
		nonces: avl.NewTree(),
	}
}

// SetNonce sets the initial nonce for an address if it hasn't been set before
func (nm *NonceManager) SetNonce(addr std.Address, nonce uint64) {
	found := nm.nonces.Has(owner.String())
	if !found {
		nm.nonces.Set(owner.String(), nonce)
	}
}

// GetNonce returns the next unused nonce for an address
func (nm *NonceManager) GetNonce(owner std.Address) uint64 {
	nonce, found := nm.nonces.Get(owner.String())
	if !found {
		panic("invalid address")
	}

	return nonce.(uint64)
}

// UseNonce consumes a nonce and returns the current value, then increments the nonce
func (nm *NonceManager) UseNonce(addr std.Address) uint64 {
	key := owner.String()

	val, found := nm.nonces.Get(key)
	if !found {
		panic("invalid address")
	}

	currentNonce := val.(uint64)

	nm.nonces.Set(key, currentNonce+1)

	return currentNonce
}

// UseCheckedNonce consumes a nonce if it matches the expected value, otherwise returns an error
func (nm *NonceManager) UseCheckedNonce(owner std.Address, nonce uint64) error {
	currentNonce := nm.UseNonce(owner)

	if nonce != currentNonce {
		return ufmt.Errorf("InvalidAccountNonce: expected %d, got %d for account %s", currentNonce, nonce, owner)
	}

	return nil
}
