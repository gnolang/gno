package main

import (
	"chain"
	"chain/runtime"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = chain.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	commondao.Invite(cross, user)

	testing.SetRealm(runtime.NewUserRealm(user))
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/demo/test"))
	dao := commondao.New(cross, "Foo")
	subDAO := commondao.NewSubDAO(cross, "Bar", dao.ID())
	subDAO.Members().Add(user)

	daoID = dao.ID()

	options := commondao.GetOptions(daoID)
	options.SetAllowMembersUpdate(false) // Disallow udpating members
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(user))

	commondao.CreateMembersUpdateProposal(cross, daoID, "", "")
}

// Error:
// forbidden
