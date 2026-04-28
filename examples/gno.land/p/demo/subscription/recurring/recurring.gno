package recurring

import (
	"chain"
	"time"

	"gno.land/p/demo/subscription"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ownable/v0"
)

// RecurringSubscription represents a subscription that requires periodic
// payments. Payment is pulled from a subscription.Sender (cap-token)
// passed by the caller — the lib does not read banker.OriginSend or
// inspect the call-chain, so it composes through any realm depth.
type RecurringSubscription struct {
	ownable.Ownable
	duration  time.Duration
	amount    int64
	payeeAddr address    // where pulled coins are sent (typically the consumer realm)
	subs      *avl.Tree // address.String() -> time.Time
}

// NewRecurringSubscription creates a new recurring subscription.
// payeeAddr is the address that receives the per-period payment; it is
// typically the consumer realm's own address.
func NewRecurringSubscription(duration time.Duration, amount int64, payeeAddr address) *RecurringSubscription {
	return &RecurringSubscription{
		Ownable:   *ownable.New(),
		duration:  duration,
		amount:    amount,
		payeeAddr: payeeAddr,
		subs:      avl.NewTree(),
	}
}

// HasValidSubscription verifies whether subscriber holds an active subscription.
func (rs *RecurringSubscription) HasValidSubscription(subscriber address) error {
	expTime, exists := rs.subs.Get(subscriber.String())
	if !exists {
		return ErrNoSub
	}
	if time.Now().After(expTime.(time.Time)) {
		return ErrSubExpired
	}
	return nil
}

// Subscribe pulls one period's worth of payment from sender into
// payeeAddr and records subscriber as having an active subscription
// until time.Now() + duration. If subscriber already has an unexpired
// subscription, returns ErrAlreadySub without pulling payment.
//
// Subscriber identity is decoupled from sender authority: a code
// realm holding a Sender can subscribe on any address's behalf
// (including itself, an EOA, or an unrelated user).
func (rs *RecurringSubscription) Subscribe(sender subscription.Sender, subscriber address) error {
	// Already-sub check first (free); avoid burning sender capacity on
	// a duplicate.
	if expTime, exists := rs.subs.Get(subscriber.String()); exists {
		if time.Now().Before(expTime.(time.Time)) {
			return ErrAlreadySub
		}
	}
	// Pull payment. If sender refuses (closed, cap exceeded, insufficient
	// funds), Send panics and the tx reverts — no subscription is
	// recorded.
	sender.Send(rs.payeeAddr, chain.NewCoins(chain.NewCoin("ugnot", rs.amount)))
	rs.subs.Set(subscriber.String(), time.Now().Add(rs.duration))
	return nil
}

// GetExpiration returns the subscription expiration for subscriber.
func (rs *RecurringSubscription) GetExpiration(subscriber address) (time.Time, error) {
	expTime, exists := rs.subs.Get(subscriber.String())
	if !exists {
		return time.Time{}, ErrNoSub
	}
	return expTime.(time.Time), nil
}

// UpdateAmount lets the subscription owner change the per-period amount.
func (rs *RecurringSubscription) UpdateAmount(newAmount int64) error {
	if !rs.Owned() {
		return ErrNotAuthorized
	}
	rs.amount = newAmount
	return nil
}

// GetAmount returns the current per-period subscription amount.
func (rs *RecurringSubscription) GetAmount() int64 {
	return rs.amount
}
