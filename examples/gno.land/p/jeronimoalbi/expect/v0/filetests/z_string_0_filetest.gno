package main

import (
	"strings"

	"gno.land/p/jeronimoalbi/expect/v0"
)

var (
	output strings.Builder
	t      = expect.MockTestingT(&output)
)

func main() {
	expect.Value(t, "foo").AsString().ToEqual("bar")
	expect.Value(t, "foo").AsString().Not().ToEqual("foo")

	println(output.String())
}

// Output:
// Expected values to match
// Got: foo
// Want: bar
// Expected values to be different
// Got: foo
