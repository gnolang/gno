// Package pausable provides a mechanism to programmatically pause and unpause
// functionality. This package allows an owner, defined via an Ownable object,
// to restrict operations or methods when the contract is in a "paused" state.
package pausable

import (
	"chain"
	"errors"

	"gno.land/p/nt/ownable/v0"
)

type Pausable struct {
	o      *ownable.Ownable
	paused bool
}

var ErrPaused = errors.New("pausable: realm is currently paused")

// NewFromOwnable is the same as New, but with a pre-existing top-level ownable
func NewFromOwnable(ownable *ownable.Ownable) *Pausable {
	return &Pausable{
		o:      ownable,
		paused: false,
	}
}

// IsPaused checks if Pausable is paused
func (p Pausable) IsPaused() bool {
	return p.paused
}

// Pause sets the state of Pausable to true, meaning all pausable functions
// are paused. rlm must be the caller's own captured cur;
// rlm.Previous().Address() must equal the underlying Ownable's owner.
func (p *Pausable) Pause(_ int, rlm realm) error {
	if !rlm.IsCurrent() {
		return ownable.ErrUnauthorized
	}
	if !p.o.OwnedBy(rlm.Previous().Address()) {
		return ownable.ErrUnauthorized
	}

	p.paused = true
	chain.Emit("Paused", "by", p.o.Owner().String())

	return nil
}

// Unpause sets the state of Pausable to false, meaning all pausable
// functions are resumed. rlm must be the caller's own captured cur;
// rlm.Previous().Address() must equal the underlying Ownable's owner.
func (p *Pausable) Unpause(_ int, rlm realm) error {
	if !rlm.IsCurrent() {
		return ownable.ErrUnauthorized
	}
	if !p.o.OwnedBy(rlm.Previous().Address()) {
		return ownable.ErrUnauthorized
	}

	p.paused = false
	chain.Emit("Unpaused", "by", p.o.Owner().String())

	return nil
}

// Ownable returns the underlying ownable
func (p *Pausable) Ownable() *ownable.Ownable {
	return p.o
}
