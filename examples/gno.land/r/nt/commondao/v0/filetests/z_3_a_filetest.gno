package main

import (
	"chain"
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner    = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user     = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	newRealm = address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
)

var (
	daoID     uint64
	realmAddr = chain.PackageAddress("gno.land/r/test")
)

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross, user)

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
	daoID = commondao.New(cross, "Foo").ID()
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))

	commondao.TransferOwnership(cross, daoID, newRealm)

	println(commondao.IsOwner(realmAddr, daoID))
	println(commondao.IsOwner(newRealm, daoID))
}

// Output:
// false
// true
