package main

import (
	"gno.land/p/aib/ibc/testing"
	"gno.land/r/aib/ibc/core"
)

// RegisterApp success
func main() {
	core.RegisterApp(cross, "app1", &testing.App{})
	core.RegisterApp(cross, "app2", &testing.App{})

	println("----------- assert render apps")
	println(core.Render("apps"))
}

// Output:
// ----------- assert render apps
// ## Registered apps
// - PortID: app1
//   - PkgPath: main
//   - Address: g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
// - PortID: app2
//   - PkgPath: main
//   - Address: g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
