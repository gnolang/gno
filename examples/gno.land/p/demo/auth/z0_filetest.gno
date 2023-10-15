package main

import (
	"std"

	"gno.land/p/demo/auth"
	"gno.land/p/demo/testutils"
)

func main() {
	test1 := testutils.TestAddress("test1")
	test2 := testutils.TestAddress("test2")
	test3 := testutils.TestAddress("test3")

	var db auth.DelegationDB

	std.TestSetOrigCaller(test1)
	db.Approve(test2, "fooperm", "barperm")
	db.Approve(test3, string(auth.PermAll)) // TODO: switch to `type any`

	{
		std.TestSetOrigCaller(test2)
		delegation := db.NewOrigDelegatedAuth(test1)
		println("test2", delegation, delegation.HasPerm("bazperm"))
	}

	{
		std.TestSetOrigCaller(test3)
		delegation := db.NewOrigDelegatedAuth(test1)
		println("test3", delegation, delegation.HasPerm("bazperm"))
	}
}

// Output:
// test2 delegated:g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7(g1w3jhxapjta047h6lta047h6lta047h6laqcyu4) false
// test3 delegated:g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7(g1w3jhxapnta047h6lta047h6lta047h6lzfhfxt) true
