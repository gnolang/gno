// Package subscription provides a basic Gno library for managing recurring
// subscriptions.
// The library allows users to create and manage different subscription models
// to control access to specific features or resources.
//
// Example Usage:
//
//	import "gno.land/p/demo/subscription"
//
//	// Duration of 30 days and amount of 1000 units subscription.New
//	var subs = subscription.NewRecurringSubscription(time.Hour*24*30, 1000)
//
//	func SomeHandler() {
//		subs.CheckOrigCaller()
//
//		// Proceed with handling the request that requires a subscription
//		// ...
//	}
package subscription // import "gno.land/p/demo/subscription"

// RecurringSubscription represents a recurring subscription model with a given
// duration and amount.
type RecurringSubscription struct {
	duration time.Duration
	amount   int64
	subs     *avl.Tree // std.Address -> time.Time
}

// NewRecurringSubscription creates a new instance of RecurringSubscription with
// the specified duration and amount.
//
// The subs parameter is an AVL tree that holds the subscription status for each
// user (std.Address) with their expiration time (time.Time).
func NewRecurringSubscription(duration time.Duration, amount int64) *RecurringSubscription {
	return &RecurringSubscription{
		duration: duration,
		amount:   amount,
		subs:     avl.NewTree(),
	}
}

// CheckOrigCaller checks the subscription status of the original caller.
//
// If the original caller does not have an active subscription or the
// subscription has expired, it panics with an error message indicating the
// required payment amount.
//
// This function should be called for each request that requires a subscription.
func (rs *RecurringSubscription) CheckOrigCaller() {
	send := std.GetOrigSend()
	caller := std.GetOrigCaller()

	// TODO implement
	good := false
	if !good {
		panic("you need to pay a subscription of " + rs.amount)
	}
}

// TODO: other subscription models
