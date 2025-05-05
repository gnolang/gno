package place

import (
	"std"

	"gno.land/p/demo/avl"
)

type User struct {
	Address std.Address
	Rank    int
	Timer   int
}

type Register struct {
	Users *avl.Tree
}

const (
	memberRank = 0
	adminRank  = 1
	timeRefill = 15
)

var (
	Registered Register
)

func init() {
	Registered.Users = avl.NewTree()
}

func RegisterMember() {
	newMember := User{
		Address: std.OriginCaller(),
		Rank:    memberRank,
		Timer:   0,
	}

	Registered.Users.Set(std.OriginCaller().String(), newMember)
}

func retrieveMember(address std.Address) (User, bool) {
	raw, ok := Registered.Users.Get(address.String())
	if !ok {
		return User{}, false
	}

	return raw.(User), true
}

func isAdmin(address std.Address) bool {
	user, ok := retrieveMember(address)
	if !ok {
		panic("User not registered")
	}

	if user.Rank != adminRank {
		return false
	}

	return true
}

func SetNewAdmin(address std.Address) {
	if !isAdmin(std.OriginCaller()) {
		panic("Must be an admin")
	}

	user, ok := retrieveMember(address)
	if !ok || user.Rank == adminRank {
		panic("User does not meet the requirement")
	}

	user.Rank = adminRank
	Registered.Users.Set(address.String(), user)
}

func DeleteUser() {
	address := std.OriginCaller()

	Registered.Users.Remove(address.String())
}

func refillTime(address std.Address) {
	user, ok := retrieveMember(address)
	if !ok {
		panic("Could not find requested user")
	}

	user.Timer = timeRefill
	Registered.Users.Set(address.String(), user)
}

func Render(path string) string {
	return "yo maaaaan"
}
