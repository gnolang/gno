package main

import (
	"std"
	"testing"
)

func main() {
	defer func() {
		println(recover())
	}()
	testing.SetRealm(std.NewCodeRealm("gno.land/p/demo/users"))
}

// Output:
// should only be called for Realms
