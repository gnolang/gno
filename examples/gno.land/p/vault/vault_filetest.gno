package main

import (
	"std"
	"time"

	"gno.land/p/grc/grc20"
	"gno.land/p/vault"
)

func main() {
	fooAdminToken := grc20.NewAdminToken("Foo", "FOO", 4)
	fooToken := fooAdminToken.GRC20()

	recoverAddr := std.Address("...")
	lockAmount := uint(1000)
	lockDuration := 5 * time.Minute
	vault := vault.New(fooToken, lockAmount, recoverAddr, lockDuration)
	println("TEST2", vault)
}

// Output:
// TEST2 &0xc000b95f80.(*gno.land/p/vault.impl)
