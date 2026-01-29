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
	expect.Value(t, 1).AsInt().ToBeLowerThan(1)
	expect.Value(t, 1).AsInt().Not().ToBeLowerThan(2)

	println(output.String())
}

// Output:
// Expected values to be lower than 1
// Got: 1
// Expected value to not to be lower than 2
// Got: 1
