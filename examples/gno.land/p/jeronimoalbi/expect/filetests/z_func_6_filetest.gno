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
	testing.SetRealm(testing.NewUserRealm(caller))

	expect.Func(t, func() { Fail(cross) }).ToCrossPanic()
	expect.Func(t, func() { Fail(cross) }).ToCrossPanic().WithMessage(msg)

	expect.Func(t, func() error { Fail(cross); return nil }).ToCrossPanic()
	expect.Func(t, func() error { Fail(cross); return nil }).ToCrossPanic().WithMessage(msg)

	expect.Func(t, func() any { Fail(cross); return nil }).ToCrossPanic()
	expect.Func(t, func() any { Fail(cross); return nil }).ToCrossPanic().WithMessage(msg)

	expect.Func(t, func() (any, error) { Fail(cross); return nil, nil }).ToCrossPanic()
	expect.Func(t, func() (any, error) { Fail(cross); return nil, nil }).ToCrossPanic().WithMessage(msg)

	expect.Func(t, func() { Success(cross) }).Not().ToCrossPanic()
	expect.Func(t, func() error { Success(cross); return nil }).Not().ToCrossPanic()
	expect.Func(t, func() any { Success(cross); return nil }).Not().ToCrossPanic()
	expect.Func(t, func() (any, error) { Success(cross); return nil, nil }).Not().ToCrossPanic()

	// None should fail, output should be empty
	print(output.String())
}

// Output:
