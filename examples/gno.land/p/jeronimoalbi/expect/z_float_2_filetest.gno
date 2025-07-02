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
	expect.Value(t, 1.2).AsFloat().ToBeGreaterOrEqualThan(1.3)
	expect.Value(t, 1.2).AsFloat().Not().ToBeGreaterOrEqualThan(1.2)

	println(output.String())
}

// Output:
// Expected values to be greater or equal than 1.3
// Got: 1.2
// Expected value to not to be greater or equal than 1.2
// Got: 1.2
