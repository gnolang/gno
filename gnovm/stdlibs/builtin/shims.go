package std

import (
	"reflect"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// TypedValue converts v to the gno.TypedValue a native binding takes in its
// place. The transpiler wraps such arguments in it.
func TypedValue(v any) gno.TypedValue {
	return gno.Go2GnoValue(nil, nil, reflect.ValueOf(v))
}

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
