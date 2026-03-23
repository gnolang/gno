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
	expect.Value(t, nil).Not().ToBeNil()
	expect.Value(t, (*int)(nil)).Not().ToBeNil()

	println(output.String())
}

// Output:
// Expected a non nil value
// Expected a non nil value
