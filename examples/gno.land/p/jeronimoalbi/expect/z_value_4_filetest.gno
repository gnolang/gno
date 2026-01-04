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
	expect.Value(t, nil).WithFailPrefix("Foo prefix").Not().ToBeNil()
	expect.Value(t, (*int)(nil)).WithFailPrefix("Foo prefix").Not().ToBeNil()

	println(output.String())
}

// Output:
// Foo prefix - Expected a non nil value
// Foo prefix - Expected a non nil value
