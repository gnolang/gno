package place

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
)

type User struct {
	Address std.Address
	Rank    int
	Timer   time.Time // Cooldown to place tiles on r/place
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

// Register a new user when clicking submit for the first time
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
		return false
	}

	return user.Rank == adminRank
}

// Give the admin rank to a new user
func SetNewAdmin(address std.Address) {
	if !isAdmin(std.PreviousRealm().Address()) {
		panic("must be an admin")
	}

	user, ok := retrieveMember(address)
	if !ok {
		panic("user does not exist")
	}

	if user.Rank == adminRank {
		panic("user is already an admin")
	}

	user.Rank = adminRank
}

// Allow user to remove itself from the registered users
func DeleteYourself() {
	address := std.PreviousRealm().Address()

	if users.Has(address.String()) {
		panic("user not registered")
	}

	if isAdmin(address) && countAdmins() <= 1 {
		panic("you are the last admin. you cannot delete yourself until more admin are nominated")
	}

	users.Remove(address.String())
}

// Ban members by removing them from the users and adding them to a blacklist
func BanMember(adr std.Address) {
	if !isAdmin(std.PreviousRealm().Address()) {
		panic("only admin can ban members")
	}

	if isAdmin(adr) {
		panic("you cannot ban an admin")
	}

	users.Remove(adr.String())
	banned.Set(adr.String(), true)
}

// Remove banned user from blacklist (need to be registered manually again)
func UnbanMember(adr std.Address) {
	if !isAdmin(std.PreviousRealm().Address()) {
		panic("only admin can ban members")
	}

	if banned.Has(adr.String()) {
		panic("user is not banned")
	}

	banned.Remove(adr.String())
}

// Set the place first admin (creator of the realm)
func setDefaultAdmin(adr std.Address) {
	newMember := &User{
		Address: adr,
		Rank:    adminRank,
		Timer:   time.Now().Add(time.Duration(-timeRefill) * time.Minute),
	}

	users.Set(adr.String(), newMember)
}
