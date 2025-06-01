package main

import (
	"std"

	"gno.land/p/demo/authctx"
	"gno.land/p/demo/testutils"
)

func main() {
	test1 := testutils.TestAddress("test1")
	test2 := testutils.TestAddress("test2")
	test3 := testutils.TestAddress("test3")
	test4 := testutils.TestAddress("test4")
	test5 := testutils.TestAddress("test5")

	var db authctx.DelegationDB

	std.TestSetOrigCaller(test1)
	var customAuthFn authctx.AuthFunc = func(role string) bool { return role == "fooperm" }
	db.Approve(test2, customAuthFn)
	db.Approve(test3, authctx.AuthorizeAll())
	db.Approve(test4, authctx.AuthorizeNone())
	db.Approve(test5, authctx.AuthorizeRoles("fooperm", "barperm"))

	println("------------- checking various AuthFunc")
	{
		std.TestSetOrigCaller(test2)
		delegation := db.NewOrigDelegatedAuthContext(test1)
		println("test2", delegation, delegation.HasRole("bazperm"), authctx.IsNative(delegation))
	}
	{
		std.TestSetOrigCaller(test3)
		delegation := db.NewOrigDelegatedAuthContext(test1)
		println("test3", delegation, delegation.HasRole("bazperm"), authctx.IsNative(delegation))
	}
	{
		std.TestSetOrigCaller(test4)
		delegation := db.NewOrigDelegatedAuthContext(test1)
		println("test4", delegation, delegation.HasRole("bazperm"), authctx.IsNative(delegation))
	}
	{
		std.TestSetOrigCaller(test5)
		delegation := db.NewOrigDelegatedAuthContext(test1)
		println("test5", delegation, delegation.HasRole("bazperm"), authctx.IsNative(delegation))
	}

	println("------------- checking IsNative")
	{
		myctx := myAuthCtxImpl{}
		println("myctx", authctx.IsNative(myctx))

		scopedMyctx := authctx.NewScopedAuthContext(myctx, authctx.AuthorizeRoles("fooperm"))
		println("scopedMyctx", authctx.IsNative(scopedMyctx))

		orig := authctx.NewOrigAuthContext()
		println("orig", authctx.IsNative(orig))

		scopedOrig := authctx.NewScopedAuthContext(orig, authctx.AuthorizeRoles("fooperm"))
		println("scopedOrig", authctx.IsNative(scopedOrig))
	}

	// TODO: test db.NewPrevDelegatedAuthContext
}

type myAuthCtxImpl struct{}

func (myAuthCtxImpl) String() string           { return "foo" }
func (myAuthCtxImpl) HasRole(role string) bool { return true }
func (myAuthCtxImpl) Addr() std.Address        { return std.Address("foobar") }

// Output:
// ------------- checking various AuthFunc
// test2 delegated:g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7/g1w3jhxapjta047h6lta047h6lta047h6laqcyu4 false true
// test3 delegated:g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7/g1w3jhxapnta047h6lta047h6lta047h6lzfhfxt true true
// test4 delegated:g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7/g1w3jhxap5ta047h6lta047h6lta047h6ldlnrjr false true
// test5 delegated:g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7/g1w3jhxap4ta047h6lta047h6lta047h6ljkuwga false true
// ------------- checking IsNative
// myctx false
// scopedMyctx false
// orig true
// scopedOrig true
