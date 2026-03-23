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
	expect.Value(t, "foo").AsString().ToHaveLength(2)
	expect.Value(t, "foo").AsString().Not().ToHaveLength(3)

	println(output.String())
}

// Output:
// Expected string length to match
// Got: 3
// Want: 2
// Expected string lengths to be different
// Got: 3
