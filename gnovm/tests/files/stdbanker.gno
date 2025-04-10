package main

import (
	"chain/runtime"
	"testing"
)

func main() {
	defer func() {
		println(recover())
	}()
	testing.SetRealm(runtime.NewCodeRealm("gno.land/p/demo/users"))
}

// Output:
// should only be called for Realms
