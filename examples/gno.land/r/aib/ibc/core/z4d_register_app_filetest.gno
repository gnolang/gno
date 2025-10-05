package main

import (
	"gno.land/p/aib/ibc/testing"
	"gno.land/r/aib/ibc/core"
)

// RegisterApp fails because portID is not alphanumeric
func main() {
	core.RegisterApp(cross, "gno.land/r/aib/ibc", &testing.App{})
}

// Error:
// invalid portID: validate port identifier: identifier gno.land/r/aib/ibc cannot contain separator '/'
