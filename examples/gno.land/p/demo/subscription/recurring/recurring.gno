package recurring

import (
	"chain"
	"time"

	"gno.land/p/jaekwon/allowancesender/v0"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ownable/v0"
)

// RecurringSubscription represents a subscription that requires periodic
// payments. Payment is pulled from a caller-supplied *AllowanceSender
// cap-token (which the constructor of this struct trusts via
// banker.IsCanonical). Funds are sent to payeeAddr, locked at
// construction so that a leaked struct pointer cannot redirect payment.
type RecurringSubscription struct {
	ownable.Ownable
	duration  time.Duration
	amount    int64
	payeeAddr address   // payment destination; set at construction.
	subs      *avl.Tree // address.String() -> time.Time
}

// NewRecurringSubscription creates a new recurring subscription. payeeAddr
// must be valid (typically the constructing realm's own address) and is
// the fixed destination for every Send pulled by Subscribe.
func NewRecurringSubscription(duration time.Duration, amount int64, payeeAddr address) *RecurringSubscription {
	if !payeeAddr.IsValid() {
		panic(ErrInvalidPayee)
	}
	return &RecurringSubscription{
		Ownable:   *ownable.New(),
		duration:  duration,
		amount:    amount,
		payeeAddr: payeeAddr,
		subs:      avl.NewTree(),
	}
}

// HasValidSubscription verifies whether subscriber holds an active
// recurring subscription.
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

// Subscribe pulls one period's worth of payment from al into payeeAddr
// and records subscriber as having an active subscription until
// time.Now() + duration. Already-subscribed-and-active returns
// ErrAlreadySub WITHOUT pulling payment. Subscriber identity is
// decoupled from cap-token authority: a single AllowanceSender holder
// can subscribe any address (gift, self, or third-party).
//
// On Send-side failure (closed allowance, cap exceeded, insufficient
// funds, etc.), al.Send panics — the whole tx reverts and no
// subscription is recorded. Atomicity covers any half-state.
func (rs *RecurringSubscription) Subscribe(al *allowancesender.AllowanceSender, subscriber address) error {
	if !subscriber.IsValid() {
		return ErrInvalidSubscriber
	}
	// Already-sub check first; avoid burning cap-token capacity on a
	// duplicate that would just be rejected.
	if expTime, exists := rs.subs.Get(subscriber.String()); exists {
		if time.Now().Before(expTime.(time.Time)) {
			return ErrAlreadySub
		}
	}
	al.Send(rs.payeeAddr, chain.NewCoins(chain.NewCoin("ugnot", rs.amount)))
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

// GetPayee returns the locked payment destination.
func (rs *RecurringSubscription) GetPayee() address {
	return rs.payeeAddr
}
