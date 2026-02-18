package main

import (
	"chain"
	"testing"

	tests "gno.land/r/tests/vm"
)

func main() {
	testing.OriginSend("gno.land/r/tests/vm", chain.Coins{{Denom: "foo", Amount: 42}, {Denom: "ugnot", Amount: 100}})
	println("tests.RealmBalance: " + tests.RealmBalance(cross))
	println("tests.RealmCoins: " + tests.RealmCoins(cross))
	println("tests.BankerOriginSend: " + tests.BankerOriginSend(cross))

	// subtests is cross-called from tests, so cur.Coins() returns empty there.
	println("tests.RTestsRealmCoins: " + tests.RTestsRealmCoins(cross))
	// banker.OriginSend() always returns the origin send regardless of call
	// depth.
	println("tests.RTestsOriginSend: " + tests.RTestsOriginSend(cross))
}

// Output:
// tests.RealmBalance: 42foo,100ugnot
// tests.RealmCoins: 42foo,100ugnot
// tests.BankerOriginSend: 42foo,100ugnot
// tests.RTestsRealmCoins:
// tests.RTestsOriginSend: 42foo,100ugnot
