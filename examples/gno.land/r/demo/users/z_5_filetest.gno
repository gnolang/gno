package main

// SEND: 200000000ugnot

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/users"
)

const admin = std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")

func main() {
	caller := testutils.TestAddress("caller")
	std.TestSetRealm(std.NewUserRealm(caller))
	std.TestSetOrigCaller(caller)
	users.Register("", "gnouser", "my profile")
	// as admin, grant invites to gnouser
	std.TestSetOrigCaller(admin)
	std.TestSetRealm(std.NewUserRealm(admin))
	users.GrantInvites(caller.String() + ":1")
	// switch back to caller
	std.TestSetOrigCaller(caller)
	std.TestSetRealm(std.NewUserRealm(caller))
	// invite another addr
	test1 := testutils.TestAddress("test1")
	users.Invite(test1.String())
	// switch to test1
	std.TestSetOrigCaller(test1)
	std.TestSetRealm(std.NewUserRealm(test1))
	std.TestSetOrigSend(std.Coins{{"dontcare", 1}}, nil)
	users.Register(caller, "satoshi", "my other profile")
	println(users.Render(""))
	println("========================================")
	println(users.Render("?page=2"))
	println("========================================")
	println(users.Render("gnouser"))
	println("========================================")
	println(users.Render("satoshi"))
	println("========================================")
	println(users.Render("badname"))
}

// Output:
//  * [archives](/r/demo/users:archives)
//  * [demo](/r/demo/users:demo)
//  * [gno](/r/demo/users:gno)
//  * [gnoland](/r/demo/users:gnoland)
//  * [gnolang](/r/demo/users:gnolang)
//  * [gnouser](/r/demo/users:gnouser)
//  * [gov](/r/demo/users:gov)
//  * [nt](/r/demo/users:nt)
//  * [satoshi](/r/demo/users:satoshi)
//  * [sys](/r/demo/users:sys)
//  * [test1](/r/demo/users:test1)
//  * [x](/r/demo/users:x)
//
//
// ========================================
//
//
// ========================================
// ## user gnouser
//
//  * address = g1vdskcmr9wf047h6lta047h6lta047h6lq5m7fe
//  * 9 invites
//
// my profile
//
// ========================================
// ## user satoshi
//
//  * address = g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7
//  * 0 invites
//  * invited by g1vdskcmr9wf047h6lta047h6lta047h6lq5m7fe
//
// my other profile
//
// ========================================
// unknown username badname
