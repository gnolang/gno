package lifetime

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
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

	if amount.AmountOf("ugnot") != ls.amount {
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

func (ls *LifetimeSubscription) ProcessPaymentGift(receiver std.Address) error {
	return ls.validatePayment(receiver)
}

func (ls *LifetimeSubscription) CheckAccess(caller std.Address) error {
	_, exists := ls.subs.Get(caller.String())

	if !exists {
		return ErrNoSub
	}

	return nil
}

func (ls *LifetimeSubscription) UpdateAmount(newAmount int64) error {
	if err := ls.CallerIsOwner(); err != nil {
		return ErrNotAuthorized
	}

	ls.amount = newAmount
	return nil
}

func (ls *LifetimeSubscription) GetAmount() int64 {
	return ls.amount
}
