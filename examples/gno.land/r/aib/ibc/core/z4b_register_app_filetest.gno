package main

import (
	"gno.land/p/aib/ibc/testing"
	"gno.land/r/aib/ibc/core"
)

// RegisterApp error when already registered
func main() {
	core.RegisterApp(cross, "testing-app", &testing.App{})
	core.RegisterApp(cross, "testing-app", &testing.App{})
}

// Error:
// route testing-app has already been registered
