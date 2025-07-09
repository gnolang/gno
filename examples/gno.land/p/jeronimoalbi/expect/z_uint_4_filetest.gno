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
	expect.Value(t, 1).AsUint().ToBeLowerOrEqualThan(0)
	expect.Value(t, 1).AsUint().Not().ToBeLowerOrEqualThan(1)

	println(output.String())
}

// Output:
// Expected values to be lower or equal than 0
// Got: 1
// Expected value to not to be lower or equal than 1
// Got: 1
