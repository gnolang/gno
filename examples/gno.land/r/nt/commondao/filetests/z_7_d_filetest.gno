package main

import (
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var parentID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross, user)

	// The origin must be the invited user
	testing.SetRealm(testing.NewUserRealm(user))

	// Create the root DAO
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/nt/commondao"))
	parentDAO := commondao.New(cross, "Parent DAO", commondao.AllowChildren(false)) // Disallow SubDAOs
	parentID = parentDAO.ID()
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/nt/commondao"))

	commondao.NewSubDAO(cross, "Foo", parentID)
}

// Error:
// SubDAO support is not enabled
