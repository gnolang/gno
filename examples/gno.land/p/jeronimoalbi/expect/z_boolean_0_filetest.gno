package main

import (
	"strings"

	"gno.land/p/jeronimoalbi/expect"
)

var (
	output strings.Builder
	t      = expect.MockTestingT(&output)
)

func main() {
	expect.Value(t, true).AsBoolean().ToEqual(false)
	expect.Value(t, true).AsBoolean().Not().ToEqual(true)

	println(output.String())
}

// Output:
// Expected values to match
// Got: true
// Want: false
// Expected values to be different
// Got: true
