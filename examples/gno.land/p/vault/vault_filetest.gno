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
	fooAdminToken.Mint(alice, 1000)
	fooToken := fooAdminToken.GRC20()

	recoverAddr := std.Address("...")
	lockAmount := uint(300)
	lockDuration := uint(5)
	v := vault.New(fooToken, lockAmount, recoverAddr, lockDuration)
	println(v)
	// TODO: example of standard flow.
}

// Output:
// from g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh to g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4 amount 300 fromBalance 1000 toBalance 0
// &0xc000c78630.(*gno.land/p/vault.impl)
