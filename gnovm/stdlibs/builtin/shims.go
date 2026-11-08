package std

import gnolang "github.com/gnolang/gno/gnovm/pkg/gnolang"

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

// NativeTypedValue is the transpiled stand-in for an argument handed to a
// native binding whose Go implementation declares that parameter as
// gnolang.TypedValue. genstd links those through with no Gno2Go conversion
// (see misc/genstd/mapping.go), so the Go signature does not spell the Gno
// parameter type and a transpiled call cannot pass its argument directly.
//
// Transpiled Go has no machine and no Gno value graph, so there is nothing
// real to hand over: the argument is still evaluated, for its side effects,
// and the binding receives the zero TypedValue. That matches what transpiled
// output already does with every other native, which it calls with a nil
// *Machine: the code type-checks and builds, it is not meant to run.
func NativeTypedValue(any) gnolang.TypedValue { return gnolang.TypedValue{} }
