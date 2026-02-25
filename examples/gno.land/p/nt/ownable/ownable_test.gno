package ownable

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
)

func TestNew(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))
	current := runtime.CurrentRealm().Address()

	o := New()
	got := o.Owner()
	uassert.Equal(t, got, current)
}

func TestNewWithOriginPanic(t *testing.T) {
	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewUserRealm(alice))

	uassert.PanicsWithMessage(t, "frame not found: cannot seek beyond origin caller override", func() {
		NewWithOrigin()
	})
}

func TestNewWithOrigin(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		// This is the only way to test crosses from a p package for now.
		o := NewWithOrigin()
		got := o.Owner()
		uassert.Equal(t, got, alice)
	})
}

func TestNewWithAddress(t *testing.T) {
	o := NewWithAddress(alice)

	got := o.Owner()
	uassert.Equal(t, got, alice)
}

func TestNewWithAddressByPrevious(t *testing.T) {
	o := NewWithAddressByPrevious(alice)

	got := o.Owner()
	uassert.Equal(t, got, alice)
}

func TestTransferOwnership(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New()
	err := o.TransferOwnership(bob)
	urequire.NoError(t, err)

	got := o.Owner()
	uassert.Equal(t, got, bob)
}

func TestTransferOwnershipByPrevious(t *testing.T) {
	var o *Ownable

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		o = NewWithOrigin()
	})

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		err := o.TransferOwnership(bob)
		urequire.NoError(t, err)
	})

	got := o.Owner()
	uassert.Equal(t, got, bob)
}

func TestTransferOwnershipUnauthorized(t *testing.T) {
	var o *Ownable

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		o = NewWithOrigin() // owned by alice, previous-mode
	})

	// Bob cannot transfer ownership.
	testing.SetRealm(testing.NewUserRealm(bob))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		uassert.ErrorContains(t, o.TransferOwnership(bob), ErrUnauthorized.Error())
	})
}

func TestDropOwnership(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New()

	err := o.DropOwnership()
	urequire.NoError(t, err, "DropOwnership failed")

	owner := o.Owner()
	uassert.Empty(t, owner, "owner should be empty")
}

func TestDropOwnershipByPrevious(t *testing.T) {
	var o *Ownable

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		o = NewWithOrigin()
	})

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		err := o.DropOwnership()
		urequire.NoError(t, err, "DropOwnership failed")
	})

	owner := o.Owner()
	uassert.Empty(t, owner, "owner should be empty")
}

func TestDropOwnershipUnauthorized(t *testing.T) {
	var o *Ownable

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		o = NewWithOrigin() // owned by alice
	})

	testing.SetRealm(testing.NewUserRealm(bob))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		uassert.ErrorContains(t, o.DropOwnership(), ErrUnauthorized.Error())
	})
}

func TestOwned(t *testing.T) {
	// Current-mode.
	testing.SetRealm(testing.NewUserRealm(alice))
	o := New()
	uassert.True(t, o.Owned())

	testing.SetRealm(testing.NewUserRealm(bob))
	uassert.False(t, o.Owned())

	// Previous-mode.
	var po *Ownable
	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		po = NewWithOrigin()
	})

	// Previous-mode Owned() requires a cross-realm context (PreviousRealm() is only
	// meaningful when called from a crossing function).
	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		uassert.True(t, po.Owned())
	})
	testing.SetRealm(testing.NewUserRealm(bob))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		uassert.False(t, po.Owned())
	})
}

func TestAssertOwned(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	o := New()

	// Should not panic.
	o.AssertOwned()

	testing.SetRealm(testing.NewUserRealm(bob))
	uassert.PanicsWithMessage(t, ErrUnauthorized.Error(), func() {
		o.AssertOwned()
	})
}

func TestErrInvalidAddress(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))

	o := New()
	err := o.TransferOwnership("")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())

	err = o.TransferOwnership("10000000001000000000100000000010000000001000000000")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
}

func TestNilReceiver(t *testing.T) {
	var o *Ownable

	owner := o.Owner()
	if owner != address("") {
		t.Errorf("expected empty address but got %v", owner)
	}

	isOwner := o.Owned()
	uassert.False(t, isOwner)

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic but got none")
		}
		if r != ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized but got %v", r)
		}
	}()
	o.AssertOwned()
}

func crossThrough(rlm runtime.Realm, cr func()) {
	testing.SetRealm(rlm)
	cr()
}
