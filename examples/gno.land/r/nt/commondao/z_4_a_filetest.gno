package main

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name  = "B"
)

var daoID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	cross(commondao.Invite)(user)

	// Create a couple of DAOs
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
	commondao.New("A")
	daoID = commondao.New(name).ID()
	commondao.New("C")
}

func main() {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))

	dao := commondao.Get(daoID)
	if dao == nil {
		panic("expected DAO to be found")
	}

	println(dao.Name() == name)
	println(dao.ID() == daoID)
}

// Output:
// true
// true
