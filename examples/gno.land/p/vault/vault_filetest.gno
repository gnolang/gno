package main

import (
	"std"
	"time"

	"gno.land/p/grc/grc20"
	"gno.land/p/testutils"
	"gno.land/p/ufmt"
	"gno.land/p/vault"
)

func main() {
	alice := testutils.TestAddress("alice")
	std.TestSetOrigCaller(alice)

	fooAdminToken := grc20.NewAdminToken("Foo", "FOO", 4)
	fooToken := fooAdminToken.GRC20()

	recoverAddr := std.Address("...")
	lockAmount := uint(1000)
	lockDuration := uint(5)
	v := vault.New(fooToken, lockAmount, recoverAddr, lockDuration)
	println(v)
	// TODO: example of standard flow.
}

// Output:
// PKG g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
// &0xc000782db0.(*gno.land/p/vault.impl)
