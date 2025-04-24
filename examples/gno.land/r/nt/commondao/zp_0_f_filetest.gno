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

var daoID uint64

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	cross(commondao.Invite)(user)

	testing.SetOriginCaller(user)
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
	dao := commondao.New("Foo")
	subDAO := commondao.NewSubDAO("Bar", dao.ID())
	subDAO.Members().Add(user)

	daoID = dao.ID()

	options := commondao.GetOptions(daoID)
	options.SetAllowMembersUpdate(false) // Disallow udpating members
}

func main() {
	testing.SetRealm(std.NewUserRealm(user))

	cross(commondao.CreateMembersUpdateProposal)(daoID, "", "")
}

// Error:
// forbidden
