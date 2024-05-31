package main

import (
	"std"
)

func main() {
	defer func() {
		println(recover())
	}()
	std.TestSetRealm(std.NewCodeRealm("gno.land/p/demo/users"))
}



// Output:
// should only be called for Realms