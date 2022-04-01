package users

import (
	"errors"
	"regexp"
	"std"

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

//----------------------------------------
// State

var name2User *avl.Tree // Name -> *User
var addr2User *avl.Tree // std.Address -> *User
var invites *avl.Tree   // string(inviter+":"+invited) -> struct{}{}
var counter int

//----------------------------------------
// Top-level functions

func other() {
	banker := std.GetBanker(std.BankerTypeTxSend)
	println(banker)
}

func Register(inviter std.Address, name string, profile string) error {
	caller := std.GetCallerAt(2)
	// assert CallTx call.
	std.AssertOriginCall()
	// assert invited or paid.
	if inviter == "" {
		// banker := std.GetBanker(std.BankerTypeTxSend)
		sent := std.GetTxSendCoins()
		// TODO: implement sent.IsGTE(...)
		if len(sent) == 1 && sent[0].Denom == "gnot" && sent[0].Amount > 2000 {
			// ok
		} else {
			return errors.New("insufficient payment")
		}
	} else {
		invitekey := string(inviter + ":" + caller)
		_, _, ok := invites.Get(invitekey)
		if !ok {
			return errors.New("invalid invitation")
		}
		invites.Remove(invitekey)
	}
	// assert not already registered.
	_, _, ok := name2User.Get(name)
	if ok {
		return errors.New("name already registered")
	}
	if caller != std.GetOrigCaller() {
		panic("should not happen")
	}
	_, _, ok = addr2User.Get(caller)
	if ok {
		return errors.New("address already registered")
	}
	// assert name is valid.
	if !reName.MatchString(name) {
		panic("invalid name: " + name)
	}
	// register.
	counter++
	user := &User{
		address: caller,
		name:    name,
		profile: profile,
		number:  counter,
	}
	name2User, _ = name2User.Set(name, user)
	addr2User, _ = addr2User.Set(caller, user)
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

func GetUserByAddress(addr string) *User {
	_, userI, ok := addr2User.Get(addr)
	if !ok {
		return nil
	}
	return userI.(*User)
}

//----------------------------------------
// Constants

var reName = regexp.MustCompile(`^[a-z]+[_a-z0-9]{5,16}$`)
