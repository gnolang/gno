package pausable

import (
	"gno.land/p/demo/ownable"
)

type Pausable struct {
	ownable.Ownable
	paused bool
}

// New returns a new Pausable struct with non-paused state as default
func New() *Pausable {
	return &Pausable{
		Ownable: ownable.New(),
		paused:  false,
	}
}

// NewFromOwnable is the same as New, but with a pre-existing top-level ownable
func NewFromOwnable(ownable ownable.Ownable) *Pausable {
	return &Pausable{
		Ownable: ownable,
		paused:  false,
	}
}

// IsPaused checks if Pausable is paused
func (p Pausable) IsPaused() bool {
	return p.paused
}

// Pause sets the state of Pausable to true, meaning all pausable functions are paused
func (p *Pausable) Pause() error {
	if !prevRealmIsOwner(p.Owner()) {
		return ErrUnauthorized
	}

	p.paused = true
	return nil
}

// Unpause sets the state of Pausable to false, meaning all pausable functions are resumed
func (p *Pausable) Unpause() error {
	if !prevRealmIsOwner(p.Owner()) {
		return ErrUnauthorized
	}

	p.paused = false
	return nil
}
