package place

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
)

type User struct {
	Address std.Address
	Rank    int
	Timer   time.Time
}

const (
	memberRank int = 0
	adminRank  int = 1

	timeRefill int = 2
)

var Users *avl.Tree = avl.NewTree()

func RegisterMember() {
	_, ok := retrieveMember(std.PreviousRealm().Address())
	if ok {
		panic("user already registered")
	}

	newMember := User{
		Address: std.PreviousRealm().Address(),
		Rank:    memberRank,
		Timer:   time.Now().Add(time.Duration(-timeRefill) * time.Minute),
	}

	Users.Set(std.PreviousRealm().Address().String(), newMember)
}

func isAdmin(address std.Address) bool {
	user, ok := retrieveMember(address)
	if !ok {
		panic("user not registered")
	}

	if user.Rank != adminRank {
		return false
	}

	return true
}

func SetNewAdmin(address std.Address) {
	if !isAdmin(std.PreviousRealm().Address()) {
		panic("must be an admin")
	}

	user, ok := retrieveMember(address)
	if !ok || user.Rank == adminRank {
		panic("user does not exist")
	}

	if user.Rank == adminRank {
		panic("user is already an admin")
	}

	user.Rank = adminRank
	Users.Set(address.String(), user)
}

func DeleteYourself() {
	address := std.PreviousRealm().Address()

	_, ok := retrieveMember(address)
	if !ok {
		panic("user not registered")
	}

	ok = isAdmin(std.PreviousRealm().Address())
	if ok && countAdmins() <= 1 {
		panic("you are the last admin. you cannot delete yourself until more admin are nominated")
	}

	Users.Remove(address.String())
}
