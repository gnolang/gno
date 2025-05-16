package guestbook

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
)

func TestSign(t *testing.T) {
	guestbook = avl.Tree{}
	hasSigned = avl.Tree{}

	testing.SetRealm(std.NewUserRealm("g1user"))
	cross(Sign)("Hello!")

	testing.SetRealm(std.NewUserRealm("g1user2"))
	cross(Sign)("Hello2!")

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
	testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoland/users/v1"))

	rec := revive(func() { cross(Sign)("Hey!") })
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

	testing.SetRealm(std.NewUserRealm("g1user"))
	cross(Sign)("Hello!")

	rec := revive(func() { cross(Sign)("Hello again!") })
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

	testing.SetRealm(std.NewUserRealm("g1user"))

	rec := revive(func() { cross(Sign)("\x00Hello!") })
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
		userAddr  std.Address = "g1user"
		adminAddr std.Address = "g1admin"
	)

	guestbook = avl.Tree{}
	hasSigned = avl.Tree{}
	owner = ownable.NewWithAddress(adminAddr)
	signatureID = 0

	testing.SetRealm(std.NewUserRealm(userAddr))

	const bad = "Very Bad Message! Nyeh heh heh!"
	cross(Sign)(bad)

	if rnd := Render(""); !strings.Contains(rnd, bad) {
		t.Fatal("render does not contain bad message", rnd)
	}

	testing.SetRealm(std.NewUserRealm(adminAddr))
	cross(AdminDelete)(signatureID.String())

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
