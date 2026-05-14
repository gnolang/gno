package lifetime

import (
	"chain"

	"gno.land/p/jaekwon/allowancesender/v0"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ownable/v0"
)

// LifetimeSubscription represents a one-time-payment subscription that
// grants permanent access. Payment is pulled from a caller-supplied
// *AllowanceSender cap-token; funds are sent to payeeAddr, which is
// locked at construction.
type LifetimeSubscription struct {
	ownable.Ownable
	amount    int64
	payeeAddr address
	subs      *avl.Tree // address.String() -> bool
}

// NewLifetimeSubscription creates a new lifetime subscription.
func NewLifetimeSubscription(amount int64, payeeAddr address) *LifetimeSubscription {
	if !payeeAddr.IsValid() {
		panic(ErrInvalidPayee)
	}
	return &LifetimeSubscription{
		Ownable:   *ownable.New(),
		amount:    amount,
		payeeAddr: payeeAddr,
		subs:      avl.NewTree(),
	}
}

// Subscribe pulls the one-time payment from al into payeeAddr and
// records subscriber as having a lifetime subscription. Already-subscribed
// returns ErrAlreadySub WITHOUT pulling payment.
func (ls *LifetimeSubscription) Subscribe(al *allowancesender.AllowanceSender, subscriber address) error {
	if !subscriber.IsValid() {
		return ErrInvalidSubscriber
	}
	if _, exists := ls.subs.Get(subscriber.String()); exists {
		return ErrAlreadySub
	}
	al.Send(ls.payeeAddr, chain.NewCoins(chain.NewCoin("ugnot", ls.amount)))
	ls.subs.Set(subscriber.String(), true)
	return nil
}

// HasValidSubscription verifies whether subscriber holds a lifetime subscription.
func (ls *LifetimeSubscription) HasValidSubscription(subscriber address) error {
	if _, exists := ls.subs.Get(subscriber.String()); !exists {
		return ErrNoSub
	}
	return nil
}

// UpdateAmount lets the subscription owner change the one-time amount.
func (ls *LifetimeSubscription) UpdateAmount(newAmount int64) error {
	if !ls.Owned() {
		return ErrNotAuthorized
	}
	ls.amount = newAmount
	return nil
}

// GetAmount returns the current one-time subscription amount.
func (ls *LifetimeSubscription) GetAmount() int64 {
	return ls.amount
}

// GetPayee returns the locked payment destination.
func (ls *LifetimeSubscription) GetPayee() address {
	return ls.payeeAddr
}
