package guestbook

import (
	"strings"
	"testing"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ownable"
)

func TestSign(t *testing.T) {
	guestbook = avl.Tree{}
	hasSigned = avl.Tree{}

	testing.SetRealm(testing.NewUserRealm("g1user"))
	Sign(cross, "Hello!")

	testing.SetRealm(testing.NewUserRealm("g1user2"))
	Sign(cross, "Hello2!")

	res := Render("")
	t.Log(res)
	if !strings.Contains(res, "> Hello!\n>\n> _Written by g1user ") {
		t.Error("does not contain first user's message")
	}
	if !strings.Contains(res, "> Hello2!\n>\n> _Written by g1user2 ") {
		t.Error("does not contain second user's message")
	}
	if guestbook.Size() != 2 {
		t.Error("invalid guestbook size")
	}
}

func TestSign_FromRealm(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1"))

	rec := revive(func() { Sign(cross, "Hey!") })
	if rec == nil {
		t.Fatal("expected panic")
	}
	recString, ok := rec.(string)
	if !ok {
		t.Fatal("not a string", rec)
	} else if recString != errNotAUser {
		t.Fatal("invalid error", recString)
	}
}

func TestSign_Double(t *testing.T) {
	// Should not allow signing twice.
	guestbook = avl.Tree{}
	hasSigned = avl.Tree{}

	testing.SetRealm(testing.NewUserRealm("g1user"))
	Sign(cross, "Hello!")

	rec := revive(func() { Sign(cross, "Hello again!") })
	if rec == nil {
		t.Fatal("expected panic")
	}
	recString, ok := rec.(string)
	if !ok {
		t.Error("type assertion failed", rec)
	} else if recString != errAlreadySigned {
		t.Error("invalid error message", recString)
	}
}

func TestSign_InvalidMessage(t *testing.T) {
	// Should not allow control characters in message.
	guestbook = avl.Tree{}
	hasSigned = avl.Tree{}

	testing.SetRealm(testing.NewUserRealm("g1user"))

	rec := revive(func() { Sign(cross, "\x00Hello!") })
	if rec == nil {
		t.Fatal("expected panic")
	}
	recString, ok := rec.(string)
	if !ok {
		t.Error("type assertion failed", rec)
	} else if recString != errInvalidCharacterInMessage {
		t.Error("invalid error message", recString)
	}
}

func TestAdminDelete(t *testing.T) {
	const (
		userAddr  address = "g1user"
		adminAddr address = "g1admin"
	)

	guestbook = avl.Tree{}
	hasSigned = avl.Tree{}
	owner = ownable.NewWithAddressByPrevious(adminAddr)
	signatureID = 0

	testing.SetRealm(testing.NewUserRealm(userAddr))

	const bad = "Very Bad Message! Nyeh heh heh!"
	Sign(cross, bad)

	if rnd := Render(""); !strings.Contains(rnd, bad) {
		t.Fatal("render does not contain bad message", rnd)
	}

	testing.SetRealm(testing.NewUserRealm(adminAddr))
	AdminDelete(cross, signatureID.String())

	if rnd := Render(""); strings.Contains(rnd, bad) {
		t.Error("render contains bad message", rnd)
	}
	if guestbook.Size() != 0 {
		t.Error("invalid guestbook size")
	}
	if hasSigned.Size() != 1 {
		t.Error("invalid hasSigned size")
	}
}
