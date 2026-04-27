package main

import (
	"chain"
	"testing"

	tests "gno.land/r/tests/vm"
)

func main() {
	testing.SetOriginSend(chain.Coins{{Denom: "foo", Amount: 42}, {Denom: "ugnot", Amount: 100}})
	println("tests.RealmSentCoins: " + tests.RealmSentCoins(cross))
	println("tests.BankerOriginSend: " + tests.BankerOriginSend(cross))

	// subtests is cross-called from tests, so cur.SentCoins() returns empty there.
	println("tests.RTestsRealmSentCoins: " + tests.RTestsRealmSentCoins(cross))
	// banker.OriginSend() always returns the origin send regardless of call
	// depth.
	println("tests.RTestsOriginSend: " + tests.RTestsOriginSend(cross))
}

// Output:
// tests.RealmSentCoins: 42foo,100ugnot
// tests.BankerOriginSend: 42foo,100ugnot
// tests.RTestsRealmSentCoins:
// tests.RTestsOriginSend: 42foo,100ugnot
