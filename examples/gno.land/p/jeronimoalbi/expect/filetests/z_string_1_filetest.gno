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
	expect.Value(t, "foo").AsString().ToBeEmpty()
	expect.Value(t, "").AsString().Not().ToBeEmpty()

	println(output.String())
}

// Output:
// Expected string to be empty
// Got: foo
// Unexpected empty string
