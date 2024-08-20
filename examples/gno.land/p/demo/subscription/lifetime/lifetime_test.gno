package lifetime

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

var (
	alice = std.Address("alice")
	bob   = std.Address("bob")
)

func TestLifetimeSubscription(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	ls := NewLifetimeSubscription(1000)

	std.TestSetOrigSend(std.Coins{{Denom: "ugnot", Amount: 1000}})
	err := ls.ProcessPayment()
	uassert.NoError(t, err, "Expected ProcessPayment to succeed")

	err = ls.CheckAccess(alice)
	uassert.NoError(t, err, "Expected Alice to have access")

	std.TestSetRealm(std.NewUserRealm(bob))
	std.TestSetOrigSend(std.Coins{{Denom: "ugnot", Amount: 1000}})
	err = ls.ProcessPaymentGift(bob)
	uassert.NoError(t, err, "Expected ProcessPaymentGift to succeed")

	err = ls.CheckAccess(bob)
	uassert.NoError(t, err, "Expected Bob to have access")

	std.TestSetRealm(std.NewUserRealm(alice))
	err = ls.UpdateAmount(2000)
	uassert.NoError(t, err, "Expected UpdateAmount to succeed")
}
