// Package subscription provides a library for managing different types of
// subscriptions in Gno applications. It supports both recurring and
// lifetime subscriptions, enabling users to access services based on their
// subscription status. This package uses a tree-based data structure to
// efficiently track and manage subscription statuses.

// Example Usage:
//
//	import "gno.land/p/demo/subscription"
//
//	Create a recurring subscription for 30 days costing 100 units.
//	recurringSub := subscription.NewRecurringSubscription(time.Hour*24*30, 100)
//
//	Create a lifetime subscription for a one-time payment of 500 units.
//	lifetimeSub := subscription.NewLifetimeSubscription(500)
//
//	func HandleRequest(caller std.Address) {
//		Check access for a recurring subscription.
//		recurringSub.CheckAccess(caller)
//
//		Perform payment for a lifetime subscription.
//		lifetimeSub.ProcessPayment(caller, 500)
//	}

package subscription

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

// Subscription interface defines standard methods that all subscription types must implement.
type Subscription interface {
	CheckAccess(std.Address)
	ProcessPayment(std.Address, int64)
}

// RecurringSubscription represents a subscription that requires periodic payments.
// It includes the duration of the subscription and the amount required per period.
type RecurringSubscription struct {
	duration time.Duration
	amount   int64
	subs     *avl.Tree // std.Address -> time.Time
}

// LifetimeSubscription represents a subscription that requires only a one-time payment.
// It grants permanent access to a service or product.
type LifetimeSubscription struct {
	amount int64
	subs   *avl.Tree // std.Address -> bool
}

// NewRecurringSubscription creates and returns a new recurring subscription.
func NewRecurringSubscription(duration time.Duration, amount int64) *RecurringSubscription {
	return &RecurringSubscription{
		duration: duration,
		amount:   amount,
		subs:     avl.NewTree(),
	}
}

// CheckAccess verifies if the caller has an active recurring subscription.
func (rs *RecurringSubscription) CheckAccess(caller std.Address) {
	expTime, exists := rs.subs.Get(caller.String())
	if !exists {
		return ErrNoSub
	}

	if time.Now().After(expTime.(time.Time)) {
		return ErrSubExpired
	}
}

// ProcessPayment processes the payment for a recurring subscription and extends its validity.
func (rs *RecurringSubscription) ProcessPayment(caller std.Address, paymentAmount int64) {
	if paymentAmount < rs.amount {
		return ErrPayment
	}

	expiration := time.Now().Add(rs.duration)
	rs.subs.Set(caller.String(), expiration)
}

// GetExpiration returns the expiration date of the recurring subscription for a given caller.
func (rs *RecurringSubscription) GetExpiration(caller std.Address) time.Time {
	expTime, exists := rs.subs.Get(caller.String())
	if !exists {
		return ErrNoSub
	}

	return expTime.(time.Time)
}

// NewLifetimeSubscription creates and returns a new lifetime subscription.
func NewLifetimeSubscription(amount int64) *LifetimeSubscription {
	return &LifetimeSubscription{
		amount: amount,
		subs:   avl.NewTree(),
	}
}

// ProcessPayment processes the payment for a lifetime subscription.
func (ls *LifetimeSubscription) ProcessPayment(caller std.Address, paymentAmount int64) {
	if paymentAmount < ls.amount {
		return ErrPayment
	}

	ls.subs.Set(caller.String(), true)
}

// CheckAccess verifies if the caller has a lifetime subscription.
func (ls *LifetimeSubscription) CheckAccess(caller std.Address) {
	_, exists := ls.subs.Get(caller.String())

	if !exists {
		return ErrNoSub
	}
}
