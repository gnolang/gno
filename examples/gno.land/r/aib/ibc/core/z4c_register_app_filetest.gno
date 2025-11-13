package main

import (
	stdtesting "testing"

	appstesting "gno.land/r/aib/ibc/apps/testing"
	"gno.land/r/aib/ibc/core"
)

// RegisterApp fail because caller is not a realm
func main() {
	stdtesting.SetRealm(stdtesting.NewUserRealm("g1user"))
	core.RegisterApp(cross, "app1", &appstesting.App{})
}

// Error:
// RegisterApp can only be called from an other realm
