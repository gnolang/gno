package std

import gno "github.com/gnolang/gno/gnovm/pkg/gnolang"

type Realm interface {
	Address() Address
	PkgPath() string
	Previous() Realm
	IsCode() bool
	IsUser() bool
	IsUserCall() bool
	IsUserRun() bool
	IsEphemeral() bool
	IsCurrent() bool
	Sub(subpath string) Realm
	Subpath() string
	String() string
}

type Address string

func (a Address) String() string { return string(a) }
func (a Address) IsValid() bool  { return false }

// TypedValueOf stands in for the raw value a native typed `any` receives, so
// transpiled calls type-check; transpiled natives are built, never run.
func TypedValueOf(any) gno.TypedValue { return gno.TypedValue{} }
