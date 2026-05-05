package main

import (
	"chain"
	"testing"

	tests "gno.land/r/tests/vm"
)

func main() {
	testing.SetOriginSend(chain.Coins{{Denom: "foo", Amount: 42}, {Denom: "ugnot", Amount: 100}})
	println("tests.BankerOriginSend: " + tests.BankerOriginSend(cross))

	// banker.OriginSend() always returns the origin send regardless of call
	// depth.
	println("tests.RTestsOriginSend: " + tests.RTestsOriginSend(cross))
}

// Output:
// tests.BankerOriginSend: 42foo,100ugnot
// tests.RTestsOriginSend: 42foo,100ugnot
