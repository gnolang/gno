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
	expect.Value(t, "foo").ToBeNil()

	println(output.String())
}

// Output:
// Expected value to be nil
// Got: foo
