// PKGPATH: gno.land/r/demo/test
package test

import (
	"strings"
	"testing"

	"gno.land/p/jeronimoalbi/expect"
)

const (
	caller = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
	msg    = "Boom!"
)

var (
	output strings.Builder
	t      = expect.MockTestingT(&output)
)

func Fail(realm) {
	panic(msg)
}

func Success(realm) {
	// No panic
}

func main() {
	wantMsg := "Tick Tock"

	testing.SetRealm(testing.NewUserRealm(caller))

	expect.Func(t, func() { Success(cross) }).ToCrossPanic()
	expect.Func(t, func() { Fail(cross) }).ToCrossPanic().WithMessage(wantMsg)

	expect.Func(t, func() error { Success(cross); return nil }).ToCrossPanic()
	expect.Func(t, func() error { Fail(cross); return nil }).ToCrossPanic().WithMessage(wantMsg)

	expect.Func(t, func() any { Success(cross); return nil }).ToCrossPanic()
	expect.Func(t, func() any { Fail(cross); return nil }).ToCrossPanic().WithMessage(wantMsg)

	expect.Func(t, func() (any, error) { Success(cross); return nil, nil }).ToCrossPanic()
	expect.Func(t, func() (any, error) { Fail(cross); return nil, nil }).ToCrossPanic().WithMessage(wantMsg)

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
