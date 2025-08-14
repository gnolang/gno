package main

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var rootID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	commondao.Invite(cross, user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(std.NewUserRealm(user))

	// Create root DAO without subDAOs
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test"))
	rootID = commondao.New("Root DAO").ID()
}

func main() {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test"))

	iter := commondao.NewIterator(rootID)

	println("Count =", iter.Count())
	println(iter.Next())
}

// Error:
// cannot directly modify readonly tainted object (w/o method): tree<VPBlock(1,0)>.node
