package main

import (
	"testing"
)

func main() {
	defer func() {
		println(recover())
	}()
	testing.SetRealm(testing.NewCodeRealm("gno.land/p/demo/users"))
}

// Output:
// should only be called for Realms
