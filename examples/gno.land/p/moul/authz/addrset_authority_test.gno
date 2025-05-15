package authz_test

import (
	"std"

	"gno.land/p/moul/authz"
)

func Example_basic() {
	// create a new addrset authority.
	// Auth is the safe-object authority implementing the Can() method and exposing Add(), Remove() and RemoveAll() that an admin can call to change the memberset.
	// privcfg is the unrestricted object allowing this contract to Add(), Remove(), and RemoveAll().
	Auth, privcfg := authz.New()
	privcfg.Add(authz.Admin, "g1admin")
	privcfg.Add("propose", "g1user1", "g1user2")
	privcfg.Add("vote", std.CurrentRealm().Address())
	// privcfg.Add("veto", externalAlreadyDefinedAuthority)

	/*
		if Auth.CanByCurrent(authz.Admin) {
			print("hello")
		}
		if Auth.Can(std.CurrentRealm().Address(), authz.Admin) {
			print("world")
		}
		Auth.AssertCurrentCan("arbitrary action")
		println("!")
		// Output: hello world!
	*/
}
