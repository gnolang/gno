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
	expect.Value(t, 1).AsInt().ToEqual(2)
	expect.Value(t, 1).AsInt().Not().ToEqual(1)

	println(output.String())
}

// Output:
// Expected values to match
// Got: 1
// Want: 2
// Expected value to be different
// Got: 1
