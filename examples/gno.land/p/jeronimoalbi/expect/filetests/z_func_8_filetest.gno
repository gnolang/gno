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

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(caller))

	expect.Func(t, func() { Fail(cross(cur)) }).Not().ToCrossPanic()
	expect.Func(t, func() { Fail(cross(cur)) }).ToCrossPanic().Not().WithMessage(msg)

	expect.Func(t, func() error { Fail(cross(cur)); return nil }).Not().ToCrossPanic()
	expect.Func(t, func() error { Fail(cross(cur)); return nil }).ToCrossPanic().Not().WithMessage(msg)

	expect.Func(t, func() any { Fail(cross(cur)); return nil }).Not().ToCrossPanic()
	expect.Func(t, func() any { Fail(cross(cur)); return nil }).ToCrossPanic().Not().WithMessage(msg)

	expect.Func(t, func() (any, error) { Fail(cross(cur)); return nil, nil }).Not().ToCrossPanic()
	expect.Func(t, func() (any, error) { Fail(cross(cur)); return nil, nil }).ToCrossPanic().Not().WithMessage(msg)

	println(output.String())
}

// Output:
// Expected func not to cross panic
// Got: Boom!
// Expected cross panic message to be different
// Got: Boom!
// Expected func not to cross panic
// Got: Boom!
// Expected cross panic message to be different
// Got: Boom!
// Expected func not to cross panic
// Got: Boom!
// Expected cross panic message to be different
// Got: Boom!
// Expected func not to cross panic
// Got: Boom!
// Expected cross panic message to be different
// Got: Boom!
