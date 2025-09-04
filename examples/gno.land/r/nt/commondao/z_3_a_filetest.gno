package main

import (
	"chain"
	"chain/runtime"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner    = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user     = chain.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	newRealm = chain.Address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
)

var (
	daoID     uint64
	realmAddr = chain.DerivePkgAddr("gno.land/r/test")
)

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(runtime.NewUserRealm(owner))
	commondao.Invite(cross, user)

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(runtime.NewUserRealm(user))
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/test"))
	daoID = commondao.New(cross, "Foo").ID()
}

func main() {
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/test"))

	commondao.TransferOwnership(cross, daoID, newRealm)

	println(commondao.IsOwner(realmAddr, daoID))
	println(commondao.IsOwner(newRealm, daoID))
}

// Output:
// false
// true
