package main

import (
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross, user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	dao := commondao.New(cross, "Foo")
	subDAO := commondao.NewSubDAO(cross, "Bar", dao.ID())
	subDAO.Members().Add(user)

	daoID = dao.ID()

	options := commondao.GetOptions(daoID)
	options.SetAllowMembersUpdate(false) // Disallow udpating members
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.CreateMembersUpdateProposal(cross, daoID, "", "")
}

// Error:
// forbidden
