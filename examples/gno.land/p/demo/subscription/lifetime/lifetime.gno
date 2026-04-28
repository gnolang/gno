package lifetime

import (
	"chain"

	"gno.land/p/demo/subscription"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ownable/v0"
)

// LifetimeSubscription represents a one-time-payment subscription that
// grants permanent access. Payment is pulled from a subscription.Sender
// passed by the caller — the lib does not read banker.OriginSend or
// inspect the call-chain.
type LifetimeSubscription struct {
	ownable.Ownable
	amount    int64
	payeeAddr address
	subs      *avl.Tree // address.String() -> bool
}

// NewLifetimeSubscription creates a new lifetime subscription.
// payeeAddr is the address that receives the one-time payment; it is
// typically the consumer realm's own address.
func NewLifetimeSubscription(amount int64, payeeAddr address) *LifetimeSubscription {
	return &LifetimeSubscription{
		Ownable:   *ownable.New(),
		amount:    amount,
		payeeAddr: payeeAddr,
		subs:      avl.NewTree(),
	}
}

// Subscribe pulls the one-time payment from sender into payeeAddr and
// records subscriber as having a lifetime subscription. If subscriber
// already has a subscription, returns ErrAlreadySub without pulling
// payment. Subscriber identity is decoupled from sender authority.
func (ls *LifetimeSubscription) Subscribe(sender subscription.Sender, subscriber address) error {
	if _, exists := ls.subs.Get(subscriber.String()); exists {
		return ErrAlreadySub
	}
	sender.Send(ls.payeeAddr, chain.NewCoins(chain.NewCoin("ugnot", ls.amount)))
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
