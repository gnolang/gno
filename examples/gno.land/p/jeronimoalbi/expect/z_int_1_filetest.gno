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
	expect.Value(t, 1).AsInt().ToBeGreaterThan(2)
	expect.Value(t, 1).AsInt().Not().ToBeGreaterThan(0)

	println(output.String())
}

// Output:
// Expected values to be gerater than 2
// Got: 1
// Expected value to not to be greater than 0
// Got: 1
