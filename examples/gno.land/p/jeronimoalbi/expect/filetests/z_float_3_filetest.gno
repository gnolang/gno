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
	expect.Value(t, 1.2).AsFloat().ToBeLowerThan(1.1)
	expect.Value(t, 1.2).AsFloat().Not().ToBeLowerThan(1.3)

	println(output.String())
}

// Output:
// Expected values to be lower than 1.1
// Got: 1.2
// Expected value to not to be lower than 1.3
// Got: 1.2
