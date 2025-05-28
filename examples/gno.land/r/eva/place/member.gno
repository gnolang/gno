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
	users  *avl.Tree = avl.NewTree()
	banned *avl.Tree = avl.NewTree()
)

func Register() {
	caller := std.PreviousRealm().Address()

	if banned.Has(caller.String()) {
		panic("you are banned")
	}

	if users.Has(caller.String()) {
		panic("user already registered")
	}

	newMember := &User{
		Address: caller,
		Rank:    memberRank,
		Timer:   time.Now().Add(time.Duration(-timeRefill) * time.Minute),
	}

	users.Set(caller.String(), newMember)
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
}

func DeleteYourself() {
	address := std.PreviousRealm().Address()

	if users.Has(address.String()) {
		panic("user not registered")
	}

	ok := isAdmin(address)
	if ok && countAdmins() <= 1 {
		panic("you are the last admin. you cannot delete yourself until more admin are nominated")
	}

	users.Remove(address.String())
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

	users.Remove(adr.String())
	banned.Set(adr.String(), true)
}

func UnbanMember(adr std.Address) {
	ok := isAdmin(std.PreviousRealm().Address())
	if !ok {
		panic("only admin can ban members")
	}

	if banned.Has(adr.String()) {
		panic("user is not banned")
	}

	banned.Remove(adr.String())
}

func setDefaultAdmin(adr std.Address) {
	newMember := &User{
		Address: adr,
		Rank:    adminRank,
		Timer:   time.Now().Add(time.Duration(-timeRefill) * time.Minute),
	}

	users.Set(adr.String(), newMember)
}
