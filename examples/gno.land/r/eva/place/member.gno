package place

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
)

// User is a struct containing the user address, rank and its cooldown timer
// User represents the members who can place tiles on r/place
type User struct {
	Address std.Address
	Rank    int
	Timer   time.Time // Cooldown timer before placing another tile
}

const (
	memberRank = iota // Default rank for user
	adminRank         // Administrator rank with privilege to manage users

	timeRefill = 2 // Cooldown in minutes
)

var (
	users  = avl.NewTree()
	banned = avl.NewTree()
)

// Register a new user when clicking submit for the first time
// Panics if the user is banned or already registered
func Register(cur realm) {
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

// SetNewAdmin promotes a member to the admin rank
// Only admin can use it
func SetNewAdmin(cur realm, address std.Address) {
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

// DeleteYourself removes the callers from the list of registered users
// Panics if caller is the last admin of the board
func DeleteYourself(cur realm) {
	addr := std.PreviousRealm().Address()

	if !users.Has(addr.String()) {
		panic("user not registered")
	}

	if isAdmin(addr) && countAdmins() <= 1 {
		panic("you are the last admin. you cannot delete yourself until more admins are nominated")
	}

	users.Remove(addr.String())
}

// BanMembers removes users from the users list and adds them to a blacklist
// Only admin can use it and they cannot ban fellow admins
func BanMember(cur realm, adr std.Address) {
	if !isAdmin(std.PreviousRealm().Address()) {
		panic("only admin can ban members")
	}

	if isAdmin(adr) {
		panic("you cannot ban an admin")
	}

	users.Remove(adr.String())
	banned.Set(adr.String(), true)
}

// UnbanMember removes a user from the blacklist
// Only admin can use it
// Previously banned user must be registered again manually
func UnbanMember(cur realm, adr std.Address) {
	if !isAdmin(std.PreviousRealm().Address()) {
		panic("only admin can unban members")
	}

	if !banned.Has(adr.String()) {
		panic("user is not banned")
	}

	banned.Remove(adr.String())
}

// isAdmin returns true if the user linked to the address is an administrator
func isAdmin(address std.Address) bool {
	user, ok := retrieveMember(address)
	if !ok {
		return false
	}

	return user.Rank == adminRank
}

// setDefaultAdmin set the first admins of the realm
func setDefaultAdmin(adr std.Address) {
	newMember := &User{
		Address: adr,
		Rank:    adminRank,
		Timer:   time.Now().Add(time.Duration(-timeRefill) * time.Minute),
	}

	users.Set(adr.String(), newMember)
}

// countAdmins counts the number of admins in r/place
func countAdmins() int {
	var adminCounter int

	users.Iterate("", "", func(_ string, val any) bool {
		user := val.(*User)

		if user.Rank == adminRank {
			adminCounter++
		}

		return false
	})

	return adminCounter
}

// retrieveMember returns a pointer to the User struct of the address
func retrieveMember(address std.Address) (*User, bool) {
	raw, ok := users.Get(address.String())
	if !ok {
		return nil, false
	}

	return raw.(*User), true
}
