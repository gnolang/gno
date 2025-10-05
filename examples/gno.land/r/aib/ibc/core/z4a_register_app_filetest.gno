package main

import (
	stdtesting "testing"

	"gno.land/p/aib/ibc/testing"
	"gno.land/r/aib/ibc/core"
)

// RegisterApp success
func main() {
	portID := core.RegisterApp(cross, &testing.App{})
	println("portID:", portID)
	stdtesting.SetRealm(stdtesting.NewCodeRealm("gno.land/r/aib/ibc/apps/xxx"))
	portID = core.RegisterApp(cross, &testing.App{})
	println("portID:", portID)

	println("----------- assert render apps")
	println(core.Render("apps"))
}

// Output:
// portID: main
// portID: gno.land_r_aib_ibc_apps_xxx
// ----------- assert render apps
// ## Registered apps
// - main
// - gno.land_r_aib_ibc_apps_xxx
