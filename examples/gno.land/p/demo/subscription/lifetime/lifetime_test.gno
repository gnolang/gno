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
	err := ls.Subscribe()
	uassert.NoError(t, err, "Expected ProcessPayment to succeed")

	err = ls.HasValidSubscription(std.PrevRealm().Addr())
	uassert.NoError(t, err, "Expected Alice to have access")
}

func TestLifetimeSubscriptionGift(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	std.TestSetOrigSend([]std.Coin{{Denom: "ugnot", Amount: 1000}}, nil)
	err := ls.GiftSubscription(bob)
	uassert.NoError(t, err, "Expected ProcessPaymentGift to succeed for Bob")

	err = ls.HasValidSubscription(bob)
	uassert.NoError(t, err, "Expected Bob to have access")

	err = ls.HasValidSubscription(charlie)
	uassert.Error(t, err, "Expected Charlie to fail access check")
}

func TestUpdateAmountAuthorization(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	err := ls.UpdateAmount(2000)
	uassert.NoError(t, err, "Expected Alice to succeed in updating amount")

	std.TestSetOrigCaller(bob)

	err = ls.UpdateAmount(3000)
	uassert.Error(t, err, "Expected Bob to fail when updating amount")
}

func TestIncorrectPaymentAmount(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	std.TestSetOrigSend([]std.Coin{{Denom: "ugnot", Amount: 500}}, nil)
	err := ls.Subscribe()
	uassert.Error(t, err, "Expected payment to fail with incorrect amount")
}

func TestMultipleSubscriptionAttempts(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	std.TestSetOrigSend([]std.Coin{{Denom: "ugnot", Amount: 1000}}, nil)
	err := ls.Subscribe()
	uassert.NoError(t, err, "Expected first subscription to succeed")

	std.TestSetOrigSend([]std.Coin{{Denom: "ugnot", Amount: 1000}}, nil)
	err = ls.Subscribe()
	uassert.Error(t, err, "Expected second subscription to fail as Alice is already subscribed")
}

func TestGiftSubscriptionWithIncorrectAmount(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	std.TestSetOrigSend([]std.Coin{{Denom: "ugnot", Amount: 500}}, nil)
	err := ls.GiftSubscription(bob)
	uassert.Error(t, err, "Expected gift subscription to fail with incorrect amount")

	err = ls.HasValidSubscription(bob)
	uassert.Error(t, err, "Expected Bob to not have access after incorrect gift subscription")
}

func TestUpdateAmountEffectiveness(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	err := ls.UpdateAmount(2000)
	uassert.NoError(t, err, "Expected Alice to succeed in updating amount")

	std.TestSetOrigSend([]std.Coin{{Denom: "ugnot", Amount: 1000}}, nil)
	err = ls.Subscribe()
	uassert.Error(t, err, "Expected subscription to fail with old amount after update")

	std.TestSetOrigSend([]std.Coin{{Denom: "ugnot", Amount: 2000}}, nil)
	err = ls.Subscribe()
	uassert.NoError(t, err, "Expected subscription to succeed with new amount")
}
