package guestbook

import (
	"strings"
	"testing"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ownable/v0"
)

func TestSign(cur realm, t *testing.T) {
	guestbook = avl.Tree{}
	hasSigned = avl.Tree{}

	testing.SetRealm(testing.NewUserRealm("g1user"))
	Sign(cross2(cur), "Hello!")

	testing.SetRealm(testing.NewUserRealm("g1user2"))
	Sign(cross2(cur), "Hello2!")

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

func TestSign_FromRealm(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users/init"))

	rec := revive(func() { Sign(cross2(cur), "Hey!") })
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

func TestSign_Double(cur realm, t *testing.T) {
	// Should not allow signing twice.
	guestbook = avl.Tree{}
	hasSigned = avl.Tree{}

	testing.SetRealm(testing.NewUserRealm("g1user"))
	Sign(cross2(cur), "Hello!")

	rec := revive(func() { Sign(cross2(cur), "Hello again!") })
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

func TestSign_InvalidMessage(cur realm, t *testing.T) {
	// Should not allow control characters in message.
	guestbook = avl.Tree{}
	hasSigned = avl.Tree{}

	testing.SetRealm(testing.NewUserRealm("g1user"))

	rec := revive(func() { Sign(cross2(cur), "\x00Hello!") })
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

func TestAdminDelete(cur realm, t *testing.T) {
	const (
		userAddr  address = "g1user"
		adminAddr address = "g1admin"
	)

	guestbook = avl.Tree{}
	hasSigned = avl.Tree{}
	owner = ownable.NewWithAddress(adminAddr)
	signatureID = 0

	testing.SetRealm(testing.NewUserRealm(userAddr))

	const bad = "Very Bad Message! Nyeh heh heh!"
	Sign(cross2(cur), bad)

	if rnd := Render(""); !strings.Contains(rnd, bad) {
		t.Fatal("render does not contain bad message", rnd)
	}

	testing.SetRealm(testing.NewUserRealm(adminAddr))
	AdminDelete(cross2(cur), signatureID.String())

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
