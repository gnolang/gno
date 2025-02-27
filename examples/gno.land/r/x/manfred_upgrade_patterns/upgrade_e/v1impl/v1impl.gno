package v1impl

import (
	"errors"

	home "gno.land/r/x/manfred_upgrade_patterns/upgrade_e"
)

var impl = &Impl{}

// init is for self-registration, but in practice, anything can register like a `maketx run` call by an admin.
func init() {
	// self register on init
	home.SetImpl(impl)
}

type Impl struct{} // make it real first

func (i Impl) Render(path string) string { return "hello from v1" }
func (i Impl) Foo() error                { return errors.New("not implemented") }
