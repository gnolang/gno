package main

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner    = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user     = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	newRealm = std.Address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
)

var (
	daoID     uint64
	realmAddr = std.DerivePkgAddr("gno.land/r/demo/commondao_test")
)

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	cross(commondao.Invite)(user)

	// Create a new DAO which gives ownership to `commondao_test`
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
	daoID = commondao.New("Foo").ID()
}

func main() {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))

	cross(commondao.TransferOwnership)(daoID, newRealm)

	println(commondao.IsOwner(realmAddr, daoID))
	println(commondao.IsOwner(newRealm, daoID))
}

// Output:
// false
// true
