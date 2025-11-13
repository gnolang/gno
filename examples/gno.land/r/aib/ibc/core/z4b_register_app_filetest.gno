package main

import (
	appstesting "gno.land/r/aib/ibc/apps/testing"
	"gno.land/r/aib/ibc/core"
)

// RegisterApp error when already registered
func main() {
	core.RegisterApp(cross, "app1", &appstesting.App{})
	core.RegisterApp(cross, "app1", &appstesting.App{})
}

// Error:
// portID app1 has already been registered
