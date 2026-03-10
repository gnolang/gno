package recurring

import (
	"chain/banker"
	"chain/runtime"
	"time"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ownable/v0"
)

// RecurringSubscription represents a subscription that requires periodic payments.
// It includes the duration of the subscription and the amount required per period.
type RecurringSubscription struct {
	ownable.Ownable
	duration time.Duration
	amount   int64
	subs     *avl.Tree // chain.Address -> time.Time
}

// NewRecurringSubscription creates and returns a new recurring subscription.
func NewRecurringSubscription(duration time.Duration, amount int64) *RecurringSubscription {
	return &RecurringSubscription{
		Ownable:  *ownable.New(),
		duration: duration,
		amount:   amount,
		subs:     avl.NewTree(),
	}
}

// HasValidSubscription verifies if the caller has an active recurring subscription.
func (rs *RecurringSubscription) HasValidSubscription(addr address) error {
	expTime, exists := rs.subs.Get(addr.String())
	if !exists {
		return ErrNoSub
	}

	if time.Now().After(expTime.(time.Time)) {
		return ErrSubExpired
	}

	return nil
}

// processSubscription processes the payment for a given receiver and renews or adds their subscription.
func (rs *RecurringSubscription) processSubscription(receiver address) error {
	amount := banker.OriginSend()

	if amount.AmountOf("ugnot") != rs.amount {
		return ErrAmt
	}

	expTime, exists := rs.subs.Get(receiver.String())

	// If the user is already a subscriber but his subscription has expired, authorize renewal
	if exists {
		expiration := expTime.(time.Time)
		if time.Now().Before(expiration) {
			return ErrAlreadySub
		}
	}

	// Renew or add subscription
	newExpiration := time.Now().Add(rs.duration)
	rs.subs.Set(receiver.String(), newExpiration)

	return nil
}

// Subscribe handles the payment for the caller's subscription.
func (rs *RecurringSubscription) Subscribe() error {
	caller := runtime.CurrentRealm().Address()

	return rs.processSubscription(caller)
}

// GiftSubscription allows the user to pay for a subscription for another user (receiver).
func (rs *RecurringSubscription) GiftSubscription(receiver address) error {
	return rs.processSubscription(receiver)
}

// GetExpiration returns the expiration date of the recurring subscription for a given caller.
func (rs *RecurringSubscription) GetExpiration(addr address) (time.Time, error) {
	expTime, exists := rs.subs.Get(addr.String())
	if !exists {
		return time.Time{}, ErrNoSub
	}

	return expTime.(time.Time), nil
}

// UpdateAmount allows the owner of the subscription contract to change the required subscription amount.
func (rs *RecurringSubscription) UpdateAmount(newAmount int64) error {
	if !rs.OwnedByCurrent() {
		return ErrNotAuthorized
	}

	rs.amount = newAmount
	return nil
}

// GetAmount returns the current amount required for each subscription period.
func (rs *RecurringSubscription) GetAmount() int64 {
	return rs.amount
}
