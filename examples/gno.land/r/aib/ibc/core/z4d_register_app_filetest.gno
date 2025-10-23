package main

import (
	appstesting "gno.land/r/aib/ibc/apps/testing"
	"gno.land/r/aib/ibc/core"
)

// RegisterApp fails because portID is not alphanumeric
func main() {
	core.RegisterApp(cross, "gno.land/r/aib/ibc", &appstesting.App{})
}

// Error:
// invalid portID: validate port identifier: identifier gno.land/r/aib/ibc cannot contain separator '/'
