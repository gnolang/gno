package lifetime

import (
	"chain"
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
)

var (
	alice   = testutils.TestAddress("alice")
	bob     = testutils.TestAddress("bob")
	charlie = testutils.TestAddress("charlie")
)

func TestLifetimeSubscription(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 1000}})
	err := ls.Subscribe()
	uassert.NoError(t, err, "Expected ProcessPayment to succeed")

	err = ls.HasValidSubscription(runtime.CurrentRealm().Address())
	uassert.NoError(t, err, "Expected Alice to have access")
}

func TestLifetimeSubscriptionGift(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 1000}})
	err := ls.GiftSubscription(bob)
	uassert.NoError(t, err, "Expected ProcessPaymentGift to succeed for Bob")

	err = ls.HasValidSubscription(bob)
	uassert.NoError(t, err, "Expected Bob to have access")

	err = ls.HasValidSubscription(charlie)
	uassert.Error(t, err, "Expected Charlie to fail access check")
}

func TestUpdateAmountAuthorization(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	err := ls.UpdateAmount(2000)
	uassert.NoError(t, err, "Expected Alice to succeed in updating amount")

	testing.SetRealm(testing.NewUserRealm(bob))

	err = ls.UpdateAmount(3000)
	uassert.Error(t, err, "Expected Bob to fail when updating amount")
}

func TestIncorrectPaymentAmount(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 500}})
	err := ls.Subscribe()
	uassert.Error(t, err, "Expected payment to fail with incorrect amount")
}

func TestMultipleSubscriptionAttempts(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 1000}})
	err := ls.Subscribe()
	uassert.NoError(t, err, "Expected first subscription to succeed")

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 1000}})
	err = ls.Subscribe()
	uassert.Error(t, err, "Expected second subscription to fail as Alice is already subscribed")
}

func TestGiftSubscriptionWithIncorrectAmount(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 500}})
	err := ls.GiftSubscription(bob)
	uassert.Error(t, err, "Expected gift subscription to fail with incorrect amount")

	err = ls.HasValidSubscription(bob)
	uassert.Error(t, err, "Expected Bob to not have access after incorrect gift subscription")
}

func TestUpdateAmountEffectiveness(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000)

	err := ls.UpdateAmount(2000)
	uassert.NoError(t, err, "Expected Alice to succeed in updating amount")

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 1000}})
	err = ls.Subscribe()
	uassert.Error(t, err, "Expected subscription to fail with old amount after update")

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 2000}})
	err = ls.Subscribe()
	uassert.NoError(t, err, "Expected subscription to succeed with new amount")
}
