package pausable

import "gno.land/p/demo/ownable"

type Pausable struct {
	paused bool
	*ownable.Ownable
}

// New returns a new Pausable struct with non-paused state as default
func New() *Pausable {
	return &Pausable{
		paused:  false,
		Ownable: ownable.New(),
	}
}

// NewFromOwnable is the same as New, but with a pre-existing top-level ownable
func NewFromOwnable(ownable *ownable.Ownable) *Pausable {
	return &Pausable{
		paused:  false,
		Ownable: ownable,
	}
}

// IsPaused checks if Pausable is paused
func (p *Pausable) IsPaused() bool {
	return p.paused
}

// SetPaused sets the state of Pausable to true, meaning all pausable functions are paused
func (p *Pausable) SetPaused() error {
	err := p.CallerIsOwner()

	if err != nil {
		return err
	}

	p.paused = true
	return nil
}

// SetUnpaused sets the state of Pausable to true, meaning all pausable functions are resumed
func (p *Pausable) SetUnpaused() error {
	err := p.CallerIsOwner()

	if err != nil {
		return err
	}

	p.paused = false
	return nil
}
