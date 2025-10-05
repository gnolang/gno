package main

import (
	"gno.land/p/aib/ibc/testing"
	"gno.land/r/aib/ibc/core"
)

// RegisterApp error when already registered
func main() {
	core.RegisterApp(cross, "app1", &testing.App{})
	core.RegisterApp(cross, "app1", &testing.App{})
}

// Error:
// portID app1 has already been registered
