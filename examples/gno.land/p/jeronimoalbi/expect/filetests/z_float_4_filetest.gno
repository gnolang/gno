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
	expect.Value(t, 1.2).AsFloat().ToBeLowerOrEqualThan(1.1)
	expect.Value(t, 1.2).AsFloat().Not().ToBeLowerOrEqualThan(1.2)

	println(output.String())
}

// Output:
// Expected values to be lower or equal than 1.1
// Got: 1.2
// Expected value to not to be lower or equal than 1.2
// Got: 1.2
