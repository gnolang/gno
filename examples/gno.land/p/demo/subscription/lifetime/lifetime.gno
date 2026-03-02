package lifetime

import (
	"chain/banker"
	"chain/runtime"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ownable"
)

// LifetimeSubscription represents a subscription that requires only a one-time payment.
// It grants permanent access to a service or product.
type LifetimeSubscription struct {
	ownable.Ownable
	amount int64
	subs   *avl.Tree // chain.Address -> bool
}

// NewLifetimeSubscription creates and returns a new lifetime subscription.
func NewLifetimeSubscription(amount int64) *LifetimeSubscription {
	return &LifetimeSubscription{
		Ownable: *ownable.New(),
		amount:  amount,
		subs:    avl.NewTree(),
	}
}

// processSubscription handles the subscription process for a given receiver.
func (ls *LifetimeSubscription) processSubscription(receiver address) error {
	amount := banker.OriginSend()

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

// Subscribe processes the payment for a lifetime subscription.
func (ls *LifetimeSubscription) Subscribe() error {
	caller := runtime.CurrentRealm().Address()
	return ls.processSubscription(caller)
}

// GiftSubscription allows the caller to pay for a lifetime subscription for another user.
func (ls *LifetimeSubscription) GiftSubscription(receiver address) error {
	return ls.processSubscription(receiver)
}

// HasValidSubscription checks if the given address has an active lifetime subscription.
func (ls *LifetimeSubscription) HasValidSubscription(addr address) error {
	_, exists := ls.subs.Get(addr.String())

	if !exists {
		return ErrNoSub
	}

	return nil
}

// UpdateAmount allows the owner of the LifetimeSubscription contract to update the subscription price.
func (ls *LifetimeSubscription) UpdateAmount(newAmount int64) error {
	if !ls.Owned() {
		return ErrNotAuthorized
	}

	ls.amount = newAmount
	return nil
}

// GetAmount returns the current subscription price.
func (ls *LifetimeSubscription) GetAmount() int64 {
	return ls.amount
}
