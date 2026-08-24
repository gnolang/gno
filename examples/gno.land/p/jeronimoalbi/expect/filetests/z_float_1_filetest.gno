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
	expect.Value(t, 1.2).AsFloat().ToBeGreaterThan(1.3)
	expect.Value(t, 1.2).AsFloat().Not().ToBeGreaterThan(1.1)

	println(output.String())
}

// Output:
// Expected values to be gerater than 1.3
// Got: 1.2
// Expected value to not to be greater than 1.1
// Got: 1.2
