package main

import (
	"std"
	"time"

	"gno.land/p/grc/grc20"
	"gno.land/p/vault"
)

func main() {
	fooToken := grc20.NewToken("Foo", "FOO", 4)

	recoverAddr := std.Address("...")
	lockAmount := 1000
	lockDuration := 5 * time.Minute
	vault := vault.New(fooToken, lockAmount, recoverAddr, lockDuration)
	println("TEST2", vault)
}

// Output:
// test
