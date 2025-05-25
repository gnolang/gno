// PKGPATH: gno.land/r/demo/test
package test

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/jeronimoalbi/expect"
)

const (
	caller = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
	msg    = "Boom!"
)

var (
	output strings.Builder
	t      = expect.MockTestingT(&output)
)

func Fail() {
	crossing()

	panic(msg)
}

func Success() {
	crossing()

	// No panic
}

func main() {
	wantMsg := "Tick Tock"

	testing.SetRealm(std.NewUserRealm(caller))

	expect.Func(t, func() { cross(Success)() }).ToCrossPanic()
	expect.Func(t, func() { cross(Fail)() }).ToCrossPanic().WithMessage(wantMsg)

	expect.Func(t, func() error { cross(Success)(); return nil }).ToCrossPanic()
	expect.Func(t, func() error { cross(Fail)(); return nil }).ToCrossPanic().WithMessage(wantMsg)

	expect.Func(t, func() any { cross(Success)(); return nil }).ToCrossPanic()
	expect.Func(t, func() any { cross(Fail)(); return nil }).ToCrossPanic().WithMessage(wantMsg)

	expect.Func(t, func() (any, error) { cross(Success)(); return nil, nil }).ToCrossPanic()
	expect.Func(t, func() (any, error) { cross(Fail)(); return nil, nil }).ToCrossPanic().WithMessage(wantMsg)

	println(output.String())
}

// Output:
// Expected function to cross panic
// Expected cross panic message to match
// Got: Boom!
// Want: Tick Tock
// Expected function to cross panic
// Expected cross panic message to match
// Got: Boom!
// Want: Tick Tock
// Expected function to cross panic
// Expected cross panic message to match
// Got: Boom!
// Want: Tick Tock
// Expected function to cross panic
// Expected cross panic message to match
// Got: Boom!
// Want: Tick Tock
