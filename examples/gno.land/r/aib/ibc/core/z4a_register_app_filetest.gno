package main

import (
	"gno.land/p/aib/ibc/testing"
	"gno.land/r/aib/ibc/core"
)

// RegisterApp success
func main() {
	core.RegisterApp(cross, "testing-app", &testing.App{})
}
