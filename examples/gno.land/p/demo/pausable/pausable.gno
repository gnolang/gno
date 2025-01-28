package pausable

import (
	"std"

	"gno.land/p/demo/ownable"
)

type Pausable struct {
	o      *ownable.Ownable
	paused bool
}

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

// Pause sets the state of Pausable to true, meaning all pausable functions are paused
func (p *Pausable) Pause() error {
	if !p.o.CallerIsOwner() {
		return ownable.ErrUnauthorized
	}

	p.paused = true
	std.Emit("Paused", "account", p.o.Owner().String())

	return nil
}

// Unpause sets the state of Pausable to false, meaning all pausable functions are resumed
func (p *Pausable) Unpause() error {
	if !p.o.CallerIsOwner() {
		return ownable.ErrUnauthorized
	}

	p.paused = false
	std.Emit("Unpaused", "account", p.o.Owner().String())

	return nil
}

// Owner returns the underlying owner of the pausable object
func (p *Pausable) Owner() std.Address {
	return p.o.Owner()
}
