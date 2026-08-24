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
	expect.Value(t, 1).AsInt().ToBeGreaterOrEqualThan(2)
	expect.Value(t, 1).AsInt().Not().ToBeGreaterOrEqualThan(1)

	println(output.String())
}

// Output:
// Expected values to be greater or equal than 2
// Got: 1
// Expected value to not to be greater or equal than 1
// Got: 1
