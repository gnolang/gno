package lifetime

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/subscription"
)

// LifetimeSubscription represents a subscription that requires only a one-time payment.
// It grants permanent access to a service or product.
type LifetimeSubscription struct {
	ownable.Ownable
	amount int64
	subs   *avl.Tree // std.Address -> bool
}

// NewLifetimeSubscription creates and returns a new lifetime subscription.
func NewLifetimeSubscription(amount int64) *LifetimeSubscription {
	return &LifetimeSubscription{
		Ownable: *ownable.New(),
		amount:  amount,
		subs:    avl.NewTree(),
	}
}

func (ls *LifetimeSubscription) validatePayment(receiver std.Address) error {
	amount := std.GetOrigSend()
	expAmt := std.Coins{{Denom: "ugnot", Amount: ls.amount}}

	if !amount.IsEqual(expAmt) {
		return ErrAmt
	}

	_, exists := ls.subs.Get(receiver.String())

	if exists {
		return ErrAlreadySub
	}

	ls.subs.Set(receiver.String(), true)

	return nil
}

// ProcessPayment processes the payment for a lifetime subscription.
func (ls *LifetimeSubscription) ProcessPayment() error {
	caller := std.PrevRealm().Addr()
	return ls.validatePayment(caller)
}

// ProcessPaymentGift processes the payement for offer a lifetime subscription
func (ls *LifetimeSubscription) ProcessPaymentGift(receiver std.Address) error {
	return ls.validatePayment(receiver)
}

// CheckAccess verifies if the caller has a lifetime subscription.
func (ls *LifetimeSubscription) CheckAccess(caller std.Address) error {
	_, exists := ls.subs.Get(caller.String())

	if !exists {
		return ErrNoSub
	}

	return nil
}

func (ls *LifetimeSubscription) UpdateAmount(newAmount int64) error {
	caller := std.PrevRealm().Addr()

	if err := ls.CallerIsOwner(); err != nil {
		return subscription.ErrNotAuthorized
	}

	ls.amount = newAmount
	return nil
}

func (ls *LifetimeSubscription) GetAmount() int64 {
	return ls.amount
}
