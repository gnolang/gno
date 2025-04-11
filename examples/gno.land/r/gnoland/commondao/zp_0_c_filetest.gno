package main

import (
	"std"
	"testing"

	"gno.land/r/gnoland/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init() {
	testing.SetOriginCaller(owner)
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/tests"))
	commondao.Invite(user)

	testing.SetOriginCaller(user)
	dao := commondao.New("Foo")
	dao.Members().Add(user)

	daoID = dao.ID()

	testing.SetRealm(std.NewUserRealm(user))
}

func main() {
	commondao.CreateMembersUpdateProposal(daoID, "not/found", "", "")
}

// Error:
// subDAO not found
