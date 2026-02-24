package main

import (
	"strconv"
	"strings"

	"gno.land/p/jeronimoalbi/expect"
)

var (
	output strings.Builder
	t      = expect.MockTestingT(&output)
)

type intStringer struct{ value int }

func (v intStringer) String() string {
	return strconv.Itoa(v.value)
}

func main() {
	expect.Value(t, false).AsBoolean().ToBeTruthy()
	expect.Value(t, true).AsBoolean().Not().ToBeTruthy()

	expect.Value(t, "FALSE").AsBoolean().ToBeTruthy()
	expect.Value(t, "TRUE").AsBoolean().Not().ToBeTruthy()

	expect.Value(t, []byte("FALSE")).AsBoolean().ToBeTruthy()
	expect.Value(t, []byte("TRUE")).AsBoolean().Not().ToBeTruthy()

	expect.Value(t, intStringer{0}).AsBoolean().ToBeTruthy()
	expect.Value(t, intStringer{1}).AsBoolean().Not().ToBeTruthy()

	println(output.String())
}

// Output:
// Expected value to be truthy
// Expected value not to be truthy
// Expected value to be truthy
// Expected value not to be truthy
// Expected value to be truthy
// Expected value not to be truthy
// Expected value to be truthy
// Expected value not to be truthy
