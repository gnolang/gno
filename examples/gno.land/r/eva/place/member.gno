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
	memberRank = iota
	adminRank

	timeRefill int = 2
)

var (
	Users  *avl.Tree = avl.NewTree()
	Banned *avl.Tree = avl.NewTree()
)

func RegisterMember() {
	address := std.PreviousRealm().Address()

	if _, banned := Banned.Get(address.String()); banned {
		panic("you are banned")
	}

	_, ok := retrieveMember(address)
	if ok {
		panic("user already registered")
	}

	newMember := User{
		Address: address,
		Rank:    memberRank,
		Timer:   time.Now().Add(time.Duration(-timeRefill) * time.Minute),
	}

	Users.Set(address.String(), newMember)
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

	ok = isAdmin(address)
	if ok && countAdmins() <= 1 {
		panic("you are the last admin. you cannot delete yourself until more admin are nominated")
	}

	Users.Remove(address.String())
}

func BanMember(adr std.Address) {
	ok := isAdmin(std.PreviousRealm().Address())
	if !ok {
		panic("only admin can ban members")
	}

	ok = isAdmin(adr)
	if ok {
		panic("you cannot ban an admin")
	}

	Users.Remove(adr.String())
	Banned.Set(adr.String(), true)
}

func UnbanMember(adr std.Address) {
	ok := isAdmin(std.PreviousRealm().Address())
	if !ok {
		panic("only admin can ban members")
	}

	_, ok = Banned.Get(adr.String())
	if !ok {
		panic("user is not banned")
	}

	Banned.Remove(adr.String())
}

func setDefaultAdmin(adr std.Address) {
	newMember := User{
		Address: adr,
		Rank:    adminRank,
		Timer:   time.Now().Add(time.Duration(-timeRefill) * time.Minute),
	}

	Users.Set(adr.String(), newMember)
}
