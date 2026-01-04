package recurring

import (
	"chain"
	"chain/runtime"
	"testing"
	"time"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
)

var (
	alice   = testutils.TestAddress("alice")
	bob     = testutils.TestAddress("bob")
	charlie = testutils.TestAddress("charlie")
)

func TestRecurringSubscription(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	rs := NewRecurringSubscription(time.Hour*24, 1000)

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 1000}})
	err := rs.Subscribe()
	uassert.NoError(t, err, "Expected ProcessPayment to succeed for Alice")

	err = rs.HasValidSubscription(runtime.CurrentRealm().Address())
	uassert.NoError(t, err, "Expected Alice to have access")

	_, err = rs.GetExpiration(runtime.CurrentRealm().Address())
	uassert.NoError(t, err, "Expected to get expiration for Alice")
}

func TestRecurringSubscriptionGift(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	rs := NewRecurringSubscription(time.Hour*24, 1000)

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 1000}})
	err := rs.GiftSubscription(bob)
	uassert.NoError(t, err, "Expected ProcessPaymentGift to succeed for Bob")

	err = rs.HasValidSubscription(bob)
	uassert.NoError(t, err, "Expected Bob to have access")

	err = rs.HasValidSubscription(charlie)
	uassert.Error(t, err, "Expected Charlie to fail access check")
}

func TestRecurringSubscriptionExpiration(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	rs := NewRecurringSubscription(time.Hour, 1000)

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 1000}})
	err := rs.Subscribe()
	uassert.NoError(t, err, "Expected ProcessPayment to succeed for Alice")

	err = rs.HasValidSubscription(runtime.CurrentRealm().Address())
	uassert.NoError(t, err, "Expected Alice to have access")

	expiration := time.Now().Add(-time.Hour * 2)
	rs.subs.Set(runtime.CurrentRealm().Address().String(), expiration)

	err = rs.HasValidSubscription(runtime.CurrentRealm().Address())
	uassert.Error(t, err, "Expected Alice's subscription to be expired")
}

func TestUpdateAmountAuthorization(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	rs := NewRecurringSubscription(time.Hour*24, 1000)

	err := rs.UpdateAmount(2000)
	uassert.NoError(t, err, "Expected Alice to succeed in updating amount")

	testing.SetRealm(testing.NewUserRealm(bob))
	err = rs.UpdateAmount(3000)
	uassert.Error(t, err, "Expected Bob to fail when updating amount")
}

func TestGetAmount(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	rs := NewRecurringSubscription(time.Hour*24, 1000)

	amount := rs.GetAmount()
	uassert.Equal(t, amount, int64(1000), "Expected the initial amount to be 1000 ugnot")

	err := rs.UpdateAmount(2000)
	uassert.NoError(t, err, "Expected Alice to succeed in updating amount")

	amount = rs.GetAmount()
	uassert.Equal(t, amount, int64(2000), "Expected the updated amount to be 2000 ugnot")
}

func TestIncorrectPaymentAmount(t *testing.T) {
	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewUserRealm(alice))
	rs := NewRecurringSubscription(time.Hour*24, 1000)

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 500}})
	err := rs.Subscribe()
	uassert.Error(t, err, "Expected payment with incorrect amount to fail")
}

func TestMultiplePaymentsForSameUser(t *testing.T) {
	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewUserRealm(alice))
	rs := NewRecurringSubscription(time.Hour*24, 1000)

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 1000}})
	err := rs.Subscribe()
	uassert.NoError(t, err, "Expected first ProcessPayment to succeed for Alice")

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 1000}})
	err = rs.Subscribe()
	uassert.Error(t, err, "Expected second ProcessPayment to fail for Alice due to existing subscription")
}

func TestRecurringSubscriptionWithMultiplePayments(t *testing.T) {
	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewUserRealm(alice))
	rs := NewRecurringSubscription(time.Hour, 1000)

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 1000}})
	err := rs.Subscribe()
	uassert.NoError(t, err, "Expected first ProcessPayment to succeed for Alice")

	err = rs.HasValidSubscription(runtime.CurrentRealm().Address())
	uassert.NoError(t, err, "Expected Alice to have access after first payment")

	expiration := time.Now().Add(-time.Hour * 2)
	rs.subs.Set(runtime.CurrentRealm().Address().String(), expiration)

	testing.SetOriginSend([]chain.Coin{{Denom: "ugnot", Amount: 1000}})
	err = rs.Subscribe()
	uassert.NoError(t, err, "Expected second ProcessPayment to succeed for Alice")

	err = rs.HasValidSubscription(runtime.CurrentRealm().Address())
	uassert.NoError(t, err, "Expected Alice to have access after second payment")
}
