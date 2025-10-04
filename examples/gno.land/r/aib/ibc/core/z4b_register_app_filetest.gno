package main

import (
	"gno.land/p/aib/ibc/testing"
	"gno.land/r/aib/ibc/core"
)

// RegisterApp error when already registered
func main() {
	core.RegisterApp(cross, &testing.App{})
	core.RegisterApp(cross, &testing.App{})
}

// Error:
// route main has already been registered
