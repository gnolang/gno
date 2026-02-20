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

func TestTransferOwnership(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New()
	err := o.TransferOwnership(bob)
	urequire.NoError(t, err)

	got := o.Owner()
	uassert.Equal(t, got, bob)
}

func TestTransferOwnershipUnauthorized(t *testing.T) {
	var o *Ownable

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		o = NewWithOrigin() // owned by alice
	})

	// Try unauthorized transfer from non-alice realm.
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		uassert.ErrorContains(t, o.TransferOwnership(alice), ErrUnauthorized.Error())
		uassert.ErrorContains(t, o.DropOwnershipByCurrent(), ErrUnauthorized.Error())
	})

	// Set realm to an unauthorized user bob.
	testing.SetRealm(testing.NewUserRealm(bob))
	// owner is alice, so bob is unauthorized.
	uassert.ErrorContains(t, o.TransferOwnership(bob), ErrUnauthorized.Error())
	uassert.ErrorContains(t, o.DropOwnershipByCurrent(), ErrUnauthorized.Error())
	// Reset realm to alice. NewWithOrigin is previous-auth only, so current-auth
	// methods stay unauthorized even for the owner.
	testing.SetRealm(testing.NewUserRealm(alice))
	uassert.ErrorContains(t, o.TransferOwnership(alice), ErrUnauthorized.Error())
	uassert.ErrorContains(t, o.DropOwnershipByCurrent(), ErrUnauthorized.Error())

	// Previous-auth actions remain authorized.
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		uassert.NoError(t, o.TransferOwnershipByPrevious(alice))
		uassert.NoError(t, o.DropOwnershipByPrevious())
	})
}

func TestTransferOwnershipByPrevious(t *testing.T) {
	var o *Ownable

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		o = NewWithOrigin() // owned by alice
	})

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		err := o.TransferOwnershipByPrevious(bob)
		urequire.NoError(t, err)
	})

	got := o.Owner()
	uassert.Equal(t, got, bob)
}

func TestTransferOwnershipByPreviousUnauthorized(t *testing.T) {
	var o *Ownable

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		o = NewWithOrigin() // owned by alice
	})

	testing.SetRealm(testing.NewUserRealm(bob))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		uassert.ErrorContains(t, o.TransferOwnershipByPrevious(bob), ErrUnauthorized.Error())
	})
}

func TestTransferOwnershipByPreviousInvalidAddress(t *testing.T) {
	var o *Ownable

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		o = NewWithOrigin() // owned by alice
	})

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		uassert.ErrorContains(t, o.TransferOwnershipByPrevious(""), ErrInvalidAddress.Error())
	})
}

func TestTransferOwnershipByPreviousCurrentModeUnauthorized(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao"))
	o := New() // owned by /r/gov/dao and current-auth only

	// Simulate /r/gov/dao crossing into another realm.
	crossThrough(testing.NewCodeRealm("gno.land/r/attacker/realm"), func() {
		uassert.ErrorContains(t, o.TransferOwnershipByPrevious(bob), ErrUnauthorized.Error())
		uassert.ErrorContains(t, o.DropOwnershipByPrevious(), ErrUnauthorized.Error())
	})
}

func TestOwnedByCurrent(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	o := New()
	uassert.True(t, o.OwnedByCurrent())
}

func TestOwnedByCurrentUnauthorized(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	o := New()

	uassert.True(t, o.OwnedByCurrent())

	unauthorizedCaller := bob
	testing.SetRealm(testing.NewUserRealm(unauthorizedCaller))
	uassert.False(t, o.OwnedByCurrent())
}

func TestOwnedByPrevious(t *testing.T) {
	var o *Ownable

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		o = NewWithOrigin()
	})

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		uassert.True(t, o.OwnedByPrevious())
		uassert.False(t, o.OwnedByCurrent())
	})
}

func TestOwnedByPreviousUnauthorized(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	var o *Ownable
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		o = NewWithOrigin()
	})

	unauthorizedCaller := bob
	testing.SetRealm(testing.NewUserRealm(unauthorizedCaller))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		uassert.False(t, o.OwnedByPrevious())
	})
}

func TestDropOwnershipByCurrent(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New()

	err := o.DropOwnershipByCurrent()
	urequire.NoError(t, err, "DropOwnership failed")

	owner := o.Owner()
	uassert.Empty(t, owner, "owner should be empty")
}

func TestErrInvalidAddress(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))

	o := New()
	err := o.TransferOwnership("")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())

	err = o.TransferOwnership("10000000001000000000100000000010000000001000000000")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
}

func TestAssertOwnedByCurrent(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New()

	// Should not panic when caller is owner
	o.AssertOwnedByCurrent()

	// Should panic when caller is not owner
	testing.SetRealm(testing.NewUserRealm(bob))
	uassert.PanicsWithMessage(t, ErrUnauthorized.Error(), func() {
		o.AssertOwnedByCurrent()
	})
}

func TestAssertOwnedByPrevious(t *testing.T) {
	var o *Ownable

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		o = NewWithOrigin()
	})

	testing.SetRealm(testing.NewUserRealm(alice))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		// Should not panic when previous is owner
		o.AssertOwnedByPrevious()
	})

	testing.SetRealm(testing.NewUserRealm(bob))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		uassert.PanicsWithMessage(t, ErrUnauthorized.Error(), func() {
			o.AssertOwnedByPrevious()
		})
	})
}

func TestOwnedModeAware(t *testing.T) {
	var prevOwned *Ownable

	testing.SetRealm(testing.NewUserRealm(alice))
	currentOwned := New()
	uassert.True(t, currentOwned.Owned())

	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		prevOwned = NewWithOrigin()
	})

	// previous-mode ownable requires previous-realm ownership.
	uassert.False(t, prevOwned.Owned())
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		uassert.True(t, prevOwned.Owned())
	})
}

func TestNilReceiver(t *testing.T) {
	var o *Ownable

	owner := o.Owner()
	if owner != address("") {
		t.Errorf("expected empty address but got %v", owner)
	}

	isOwner := o.OwnedByPrevious()
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
	o.AssertOwnedByPrevious()
}

func crossThrough(rlm runtime.Realm, cr func()) {
	testing.SetRealm(rlm)
	cr()
}
