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
// [{"port_id":"app1","pkg_path":"main","address":"g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4"},{"port_id":"app2","pkg_path":"main","address":"g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4"}]
