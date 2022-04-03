package users

import (
	"regexp"
	"std"
	"strconv"
	"strings"

	"gno.land/p/avl"
)

//----------------------------------------
// Types

type User struct {
	address std.Address
	name    string
	profile string
	number  int
	invites int
	inviter std.Address
}

func (u *User) Render() string {
	return "## user " + u.name + " (" + string(u.address) + ")," +
		"invites:" + strconv.Itoa(u.invites) + "," +
		"inviter:" + string(u.inviter) + "\n\n" +
		u.profile + "\n" // XXX make separate page; or quote each line.
}

//----------------------------------------
// State

var admin std.Address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"
var name2User *avl.Tree // Name -> *User
var addr2User *avl.Tree // std.Address -> *User
var invites *avl.Tree   // string(inviter+":"+invited) -> true
var counter int

//----------------------------------------
// Top-level functions

func Register(inviter std.Address, name string, profile string) {
	// assert CallTx call.
	std.AssertOriginCall()
	// assert invited or paid.
	caller := std.GetCallerAt(2)
	if caller != std.GetOrigCaller() {
		panic("should not happen") // because std.AssertOrigCall().
	}
	if inviter == "" {
		// banker := std.GetBanker(std.BankerTypeTxSend)
		sent := std.GetTxSendCoins()
		// TODO: implement sent.IsGTE(...)
		if len(sent) == 1 && sent[0].Denom == "gnot" && sent[0].Amount == 2000 {
			// ok
		} else {
			panic("payment must be exactly 2000 gnots")
		}
	} else {
		invitekey := string(inviter + ":" + caller)
		_, _, ok := invites.Get(invitekey)
		if !ok {
			panic("invalid invitation")
		}
		invites.Remove(invitekey)
	}
	// assert not already registered.
	_, _, ok := name2User.Get(name)
	if ok {
		panic("name already registered")
	}
	_, _, ok = addr2User.Get(caller.String())
	if ok {
		panic("address already registered")
	}
	// assert name is valid.
	if !reName.MatchString(name) {
		panic("invalid name: " + name + " (must be at least 6 characters, lowercase alphanumeric with underscore)")
	}
	// register.
	counter++
	user := &User{
		address: caller,
		name:    name,
		profile: profile,
		number:  counter,
		inviter: inviter,
	}
	name2User, _ = name2User.Set(name, user)
	addr2User, _ = addr2User.Set(caller.String(), user)
}

func Invite(invitee string) {
	// assert CallTx call.
	std.AssertOriginCall()
	// get caller/inviter.
	caller := std.GetCallerAt(2)
	if caller != std.GetOrigCaller() {
		panic("should not happen") // because std.AssertOrigCall().
	}
	lines := strings.Split(invitee, "\n")
	if caller == admin {
		// nothing to do, all good
	} else {
		// ensure has invites.
		_, userI, ok := addr2User.Get(caller.String())
		if !ok {
			panic("user unknown")
		}
		user := userI.(*User)
		if user.invites <= 0 {
			panic("user has no invite tokens")
		}
		user.invites -= len(lines)
		if user.invites < 0 {
			panic("user has insufficient invite tokens")
		}
	}
	// for each line...
	for _, line := range lines {
		// record invite.
		invitekey := string(caller) + ":" + string(line)
		invites, _ = invites.Set(invitekey, true)
	}
}

func GrantInvites(invites string) {
	// assert CallTx call.
	std.AssertOriginCall()
	// assert admin.
	caller := std.GetCallerAt(2)
	if caller != std.GetOrigCaller() {
		panic("should not happen") // because std.AssertOrigCall().
	}
	if caller != admin {
		panic("unauthorized")
	}
	// for each line...
	lines := strings.Split(invites, "\n")
	for _, line := range lines {
		// parse addr and invites.
		var addr string
		var invites int
		parts := strings.Split(line, ":")
		if len(parts) == 1 { // short for :1.
			addr = parts[0]
			invites = 1
		} else if len(parts) == 2 {
			addr = parts[0]
			invites_, err := strconv.Atoi(parts[1])
			if err != nil {
				panic(err)
			}
			invites = int(invites_)
		} else {
			panic("should not happen")
		}
		// give invites.
		_, userI, ok := addr2User.Get(addr)
		if !ok {
			panic("invalid user")
		}
		user := userI.(*User)
		user.invites += invites
	}
}

//----------------------------------------
// Exposed public functions

func GetUserByName(name string) *User {
	_, userI, ok := name2User.Get(name)
	if !ok {
		return nil
	}
	return userI.(*User)
}

func GetUserByAddress(addr std.Address) *User {
	_, userI, ok := addr2User.Get(addr.String())
	if !ok {
		return nil
	}
	return userI.(*User)
}

//----------------------------------------
// Constants

var reName = regexp.MustCompile(`^[a-z]+[_a-z0-9]{5,16}$`)

//----------------------------------------
// Render main page

func Render(path string) string {
	doc := ""
	name2User.Iterate("", "", func(t *avl.Tree) bool {
		value := t.Value()
		doc += value.(*User).Render() + "\n"
		return false
	})
	return doc
}
