package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	perms := boards2.GetRealmPermissions()
	println(perms != nil)
}

// Output:
// true
