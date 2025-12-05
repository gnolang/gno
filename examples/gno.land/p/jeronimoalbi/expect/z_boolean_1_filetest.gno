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
	expect.Value(t, true).AsBoolean().ToBeFalsy()
	expect.Value(t, false).AsBoolean().Not().ToBeFalsy()

	expect.Value(t, "TRUE").AsBoolean().ToBeFalsy()
	expect.Value(t, "FALSE").AsBoolean().Not().ToBeFalsy()

	expect.Value(t, []byte("TRUE")).AsBoolean().ToBeFalsy()
	expect.Value(t, []byte("FALSE")).AsBoolean().Not().ToBeFalsy()

	expect.Value(t, intStringer{1}).AsBoolean().ToBeFalsy()
	expect.Value(t, intStringer{0}).AsBoolean().Not().ToBeFalsy()

	println(output.String())
}

// Output:
// Expected value to be falsy
// Expected value not to be falsy
// Expected value to be falsy
// Expected value not to be falsy
// Expected value to be falsy
// Expected value not to be falsy
// Expected value to be falsy
// Expected value not to be falsy
