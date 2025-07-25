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
	expect.Value(t, 1).AsString()
	expect.Value(t, 1).AsBoolean()
	expect.Value(t, 1).AsFloat()
	expect.Value(t, 1).AsUint()
	expect.Value(t, "").AsInt()

	println(output.String())
}

// Output:
// Failed: incompatible type: expected a string value
// Got: int
// Failed: incompatible type: expected a boolean value
// Got: int
// incompatible type: expected a float value
// Got: int
// Failed: incompatible type: expected an int value
// Got: string
