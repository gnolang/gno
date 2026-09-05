package grc20xfer

import "gno.land/r/demo/defi/foo20"

// TokenID returns the identifier of the FOO token used by this realm.
func TokenID() string {
	return foo20.Token.ID()
}

// Faucet requests unrestricted FOO for this realm.
func Faucet(cur realm) {
	foo20.Faucet(cross(cur))
}

// Transfer sends FOO held by this realm.
func Transfer(cur realm, to address, amount int64) {
	if err := foo20.Token.RealmTeller(0, cur).Transfer(0, cur, to, amount); err != nil {
		panic(err)
	}
}
