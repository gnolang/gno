package ownable

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

var (
	alice          = testutils.TestAddress("alice")
	aliceUserRealm = std.NewUserRealm(alice)
	aliceCodeRealm = std.NewCodeRealm("gno.land/r/alice/realm")

	bob          = testutils.TestAddress("bob") // g1vfhkyh6lta047h6lta047h6lta047h6l03vdhu
	bobUserRealm = std.NewUserRealm(bob)
	bobCodeRealm = std.NewCodeRealm("gno.land/r/bob/realm")
)

func TestNewWithCurrent(t *testing.T) {
	printStack()

	testing.SetRealm(aliceCodeRealm)

	crossThrough(aliceCodeRealm, func() {
		o := NewWithCurrent()
		got := o.Owner()
		uassert.Equal(t, got, aliceCodeRealm.Address())
	})
}

func TestNewWithPrevious(t *testing.T) {
	testing.SetRealm(aliceUserRealm) // sets user realm + origin caller

	crossThrough(aliceCodeRealm, func() {
		o := NewWithPrevious()
		got := o.Owner()
		uassert.Equal(t, got, alice)
	})
}

func TestNewWithOrigin(t *testing.T) {
	testing.SetRealm(aliceUserRealm)

	// Through a realm
	crossThrough(aliceCodeRealm, func() {
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

func TestTransferOwnershipByPrevious(t *testing.T) {
	testing.SetRealm(aliceUserRealm)

	crossThrough(aliceCodeRealm, func() {
		o := NewWithPrevious()
		uassert.NoError(t, o.TransferOwnershipByPrevious(bob))
		uassert.Equal(t, o.Owner(), bob)
	})
}

func TestOwnedByPreviousUnauthorized(t *testing.T) {
	testing.SetRealm(aliceUserRealm)

	var o *Ownable
	crossThrough(aliceCodeRealm, func() {
		o = NewWithPrevious()
		urequire.Equal(t, o.Owner(), alice)
	})

	unauthorized := bobUserRealm
	testing.SetRealm(unauthorized)
	crossThrough(aliceCodeRealm, func() {
		uassert.Error(t, o.TransferOwnershipByPrevious(bob))
	})
}

func TestTransferOwnershipByCurrent(t *testing.T) {
	testing.SetRealm(aliceUserRealm)

	crossThrough(aliceCodeRealm, func() {
		o := NewWithCurrent()

	})
}

//func TestTransferOwnershipByCurrentUnauthorized(t *testing.T) {
//	var o *Ownable
//
//	testing.SetRealm(aliceUserRealm)
//	o = NewWithCurrent()
//
//	// Try unauthorized transfer from non-alice realm.
//	crossThrough(std.NewCodeRealm("gno.land/r/test/test"), func() {
//		uassert.ErrorContains(t, o.TransferOwnershipByCurrent(alice), ErrUnauthorized.Error())
//		uassert.ErrorContains(t, o.DropOwnershipByCurrent(), ErrUnauthorized.Error())
//	})
//
//	// Set realm to an unauthorized user bob.
//	testing.SetRealm(std.NewUserRealm(bob))
//	// current is gno.land/r/test/test so of course errors.
//	uassert.ErrorContains(t, o.TransferOwnership(bob), ErrUnauthorized.Error())
//	uassert.ErrorContains(t, o.DropOwnershipByCurrent(), ErrUnauthorized.Error())
//	// Reset realm to alice.
//	testing.SetRealm(aliceUserRealm)
//	uassert.NoError(t, o.TransferOwnership(alice))
//	uassert.NoError(t, o.DropOwnershipByCurrent())
//}

//	func TestOwnedByCurrent(t *testing.T) {
//		testing.SetRealm(aliceUserRealm)
//		o := New()
//		uassert.True(t, o.OwnedByCurrent())
//	}
//
//	func TestOwnedByCurrentUnauthorized(t *testing.T) {
//		testing.SetOriginCaller(alice)
//		testing.SetRealm(aliceUserRealm)
//
//		var o *Ownable
//		crossThrough(std.NewCodeRealm("gno.land/r/test/test"), func() {
//			o = NewWithOrigin()
//		})
//
//		uassert.True(t, o.OwnedByCurrent())
//
//		unauthorizedCaller := bob
//		testing.SetRealm(std.NewUserRealm(unauthorizedCaller))
//		uassert.False(t, o.OwnedByCurrent())
//	}
func TestOwnedByPrevious(t *testing.T) {
	testing.SetRealm(aliceUserRealm)

	crossThrough(aliceCodeRealm, func() {
		o := NewWithPrevious()
		uassert.True(t, o.OwnedByPrevious())
	})
}

func TestDropOwnershipByCurrent(t *testing.T) {
	testing.SetRealm(aliceUserRealm)

	crossThrough(aliceCodeRealm, func() {
		o := NewWithCurrent()
		urequire.NoError(t, o.DropOwnershipByCurrent(), "DropOwnership failed")
		owner := o.Owner()
		uassert.Empty(t, owner, "owner should be empty")
	})
}

func TestErrInvalidAddress(t *testing.T) {
	testing.SetRealm(aliceUserRealm)

	crossThrough(aliceCodeRealm, func() {
		o := NewWithPrevious()
		err := o.TransferOwnershipByPrevious("")
		uassert.ErrorContains(t, err, ErrInvalidAddress.Error())

		err = o.TransferOwnershipByPrevious("10000000001000000000100000000010000000001000000000")
		uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
	})
}

func TestAssertOwnedByCurrent(t *testing.T) {
	testing.SetRealm(aliceUserRealm)

	o := NewWithCurrent()

	// Should not panic when caller is owner
	o.AssertOwnedByCurrent()

	// Should panic when caller is not owner
	testing.SetRealm(std.NewUserRealm(bob))
	uassert.PanicsWithMessage(t, ErrUnauthorized.Error(), func() {
		o.AssertOwnedByCurrent()
	})
}

func TestAssertOwnedByPrevious(t *testing.T) {
	testing.SetRealm(aliceUserRealm)

	crossThrough(aliceCodeRealm, func() {
		o := NewWithPrevious()
		// Should not panic when previous is owner
		o.AssertOwnedByPrevious()

		// Should panic when previous is not owner
		testing.SetRealm(std.NewUserRealm(bob))
		uassert.PanicsWithMessage(t, ErrUnauthorized.Error(), func() {
			o.AssertOwnedByCurrent()
		})
	})
}

func TestNilReceiver(t *testing.T) {
	var o *Ownable

	owner := o.Owner()
	if owner != std.Address("") {
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

func crossThrough(rlm std.Realm, cr func()) {
	testing.SetRealm(rlm)
	cr()
}

func printStack() {
	println("Origin: " + std.OriginCaller())
	println("Previous: " + std.PreviousRealm().Address())
	println("Current: " + std.CurrentRealm().Address())
}
