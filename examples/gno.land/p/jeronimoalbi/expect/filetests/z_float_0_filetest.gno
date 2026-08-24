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
	expect.Value(t, 1.2).AsFloat().ToEqual(1.1)
	expect.Value(t, 1.2).AsFloat().Not().ToEqual(1.2)

	println(output.String())
}

// Output:
// Expected values to match
// Got: 1.2
// Want: 1.1
// Expected value to be different
// Got: 1.2
