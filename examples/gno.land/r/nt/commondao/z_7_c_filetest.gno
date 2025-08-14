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

var parentID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	commondao.Invite(cross, user)

	// The origin must be the invited user
	testing.SetRealm(std.NewUserRealm(user))

	// Create the root DAO
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test"))
	parentID = commondao.New("Parent DAO").ID()
}

func main() {
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test"))

	commondao.NewSubDAO("", parentID)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): tree<VPBlock(1,0)>.node
