package main

import (
	stdtesting "testing"

	"gno.land/p/aib/ibc/testing"
	"gno.land/r/aib/ibc/core"
)

// RegisterApp fail because caller is not a realm
func main() {
	stdtesting.SetRealm(stdtesting.NewUserRealm("g1user"))
	core.RegisterApp(cross, "app1", &testing.App{})
}

// Error:
// RegisterApp can only be called from an other realm
