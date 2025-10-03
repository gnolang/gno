package main

import (
	stdtesting "testing"

	"gno.land/p/aib/ibc/testing"
	"gno.land/r/aib/ibc/core"
)

// RegisterApp success
func main() {
	core.RegisterApp(cross, &testing.App{})
	stdtesting.SetRealm(stdtesting.NewCodeRealm("gno.land/r/aib/ibc/transfer"))
	core.RegisterApp(cross, &testing.App{})

	println(core.Render("apps"))
}

// Output:
// ## Registered apps
// - main
// - gno.land|r|aib|ibc|transfer
