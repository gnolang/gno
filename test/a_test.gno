package main

import (
	"std"
	"testing"
)

func TestHello(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm("gno.land/r/demo/users"))
	Call()
}
