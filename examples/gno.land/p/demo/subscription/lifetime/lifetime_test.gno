package lifetime

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var (
	alice   = testutils.TestAddress("alice")
	bob     = testutils.TestAddress("bob")
	charlie = testutils.TestAddress("charlie")
)

func TestLifetimeSubscription(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	std.TestSetOrigSend([]std.Coin{{Denom: "ugnot", Amount: 1000}}, nil)
	err := ls.ProcessPayment()
	uassert.NoError(t, err, "Expected ProcessPayment to succeed")

	err = ls.CheckAccess(std.PrevRealm().Addr())
	uassert.NoError(t, err, "Expected Alice to have access")
}

func TestLifetimeSubscriptionGift(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	std.TestSetOrigSend([]std.Coin{{Denom: "ugnot", Amount: 1000}}, nil)
	err := ls.ProcessPaymentGift(bob)
	uassert.NoError(t, err, "Expected ProcessPaymentGift to succeed for Bob")

	err = ls.CheckAccess(bob)
	uassert.NoError(t, err, "Expected Bob to have access")

	err = ls.CheckAccess(charlie)
	uassert.Error(t, err, "Expected Charlie to fail access check")
}

func TestUpdateAmountAuthorization(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	err := ls.UpdateAmount(2000)
	uassert.NoError(t, err, "Expected Alice to succeed in updating amount")

	std.TestSetRealm(std.NewUserRealm(bob))

	err = ls.UpdateAmount(3000)
	uassert.Error(t, err, "Expected Bob to fail when updating amount")
}
