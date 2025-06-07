// Package userbook demonstrates a small userbook system working with gnoweb
package userbook

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

type Signup struct {
	address_XXX std.Address
	ordinal     int
	timestamp   time.Time
}

var (
	signupsTree = avl.NewTree()
	tracker     = avl.NewTree()
	idCounter   seqid.ID
)

const signUpEvent = "SignUp"

func init() {
	SignUp(cross) // Sign up the deployer
}

func SignUp(cur realm) string {
	// Get transaction caller
	caller := std.PreviousRealm().Address()

	// Check if the user is already signed up
	if _, exists := tracker.Get(caller.String()); exists {
		panic(caller.String() + " is already signed up!")
	}

	now := time.Now()

	// Sign up the user
	signupsTree.Set(idCounter.Next().String(), &Signup{
		caller,
		signupsTree.Size(),
		now,
	})

	tracker.Set(caller.String(), struct{}{})

	std.Emit(signUpEvent, "account", caller.String())

	return ufmt.Sprintf("%s added to userbook! Timestamp: %s", caller.String(), now.Format(time.RFC822Z))
}
