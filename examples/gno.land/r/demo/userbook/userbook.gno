// Package userbook demonstrates a small userbook system working with gnoweb
package userbook

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type Signup struct {
	address   std.Address
	ordinal   int
	timestamp time.Time
}

var signupsTree = avl.NewTree()

const signUpEvent = "SignUp"

func init() {
	SignUp() // Sign up the deployer
}

func SignUp() string {
	// Get transaction caller
	caller := std.PrevRealm().Addr()

	// Check if the user is already signed up
	if _, exists := signupsTree.Get(caller.String()); exists {
		panic(caller + " is already signed up!")
	}

	now := time.Now()
	// Sign up the user
	signupsTree.Set(caller.String(), &Signup{
		std.PrevRealm().Addr(),
		signupsTree.Size(),
		now,
	})

	std.Emit(signUpEvent, "SignedUpAccount", caller.String())

	return ufmt.Sprintf("%s added to userbook! Timestamp: %s", caller.String(), now.Format(time.RFC822Z))
}
