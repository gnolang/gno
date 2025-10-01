// Package userbook demonstrates a small userbook system working with gnoweb
package userbook

import (
	"chain"
	"chain/runtime"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"
	"gno.land/p/nt/ufmt"
)

type Signup struct {
	address_XXX address
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
	caller := runtime.PreviousRealm().Address()

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

	chain.Emit(signUpEvent, "account", caller.String())

	return ufmt.Sprintf("%s added to userbook! Timestamp: %s", caller.String(), now.Format(time.RFC822Z))
}
