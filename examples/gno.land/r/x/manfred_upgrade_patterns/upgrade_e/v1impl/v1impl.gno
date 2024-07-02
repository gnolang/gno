package v1impl

import (
	"errors"

	home "gno.land/r/x/manfred_upgrade_patterns/upgrade_e"
)

// init is for self-registration, but in practice, anything can register like a `maketx run` call by an admin.
func init() {
	// self register on init
	impl := &Impl{}
	home.SetImpl(impl)
}

type Impl struct{}

func (i Impl) Render(path string) string { return "hello from v1" }
func (i Impl) Foo() error                { return errors.New("not implemented") }
