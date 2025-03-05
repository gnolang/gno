package main

import (
	"std"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var daoID uint64

func init() {
	std.TestSetOriginCaller(owner)
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/tests"))
	commondao.Invite(user)

	std.TestSetOriginCaller(user)
	dao := commondao.New("Foo")
	subDAO := commondao.NewSubDAO("Bar", "bar", dao)
	subDAO.Members().Add(user)

	daoID = dao.ID()

	std.TestSetRealm(std.NewUserRealm(user))
}

func main() {
	commondao.CreateMembersUpdateProposal(daoID, "bar", "not-an-address", "")
}

// Error:
// invalid address: not-an-address
