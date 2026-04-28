// Package subscription provides a flexible system for managing both
// recurring and lifetime subscriptions in Gno applications. Payment is
// pulled through a Sender (cap-token); the library does not read
// banker.OriginSend or inspect the call-chain, so subscriptions compose
// through any realm depth without an EOA-only restriction. Any
// caller holding a Sender bound to sufficient funds can subscribe —
// itself, an EOA, or any other address — and the recorded subscriber
// identity is decoupled from who pays.
//
// Subscriptions are tracked using an AVL tree.
//
// Usage:
//
// Import a sub-package (`recurring` or `lifetime`) to instantiate a
// subscription instance. The consumer realm constructs the subscription
// with its own address as the payee, and exposes a wrapper function
// that builds a Sender (e.g. via gno.land/p/jaekwon/allowancesender/v0)
// and forwards it to Subscribe.
//
// Recurring Subscription:
//
// Recurring subscriptions require periodic payments to maintain access.
//
// Example:
//
//	// In the consumer realm:
//	var recSub = recurring.NewRecurringSubscription(
//	    time.Hour*24*30,                       // 30-day periods
//	    100,                                   // 100 ugnot per period
//	    runtime.CurrentRealm().Address(),      // pay this realm
//	)
//
//	// Wrapper: caller-supplied AllowanceSender authorizes the per-period
//	// pull. The realm decides who is recorded as the subscriber.
//	func Subscribe(_ realm, sender subscription.Sender, subscriber address) {
//	    if err := recSub.Subscribe(sender, subscriber); err != nil {
//	        panic(err)
//	    }
//	}
//
// Lifetime Subscription:
//
// Lifetime subscriptions require a one-time payment for permanent access.
//
// Example:
//
//	var lifeSub = lifetime.NewLifetimeSubscription(
//	    500,                              // 500 ugnot one-time fee
//	    runtime.CurrentRealm().Address(), // pay this realm
//	)
//
//	func Subscribe(_ realm, sender subscription.Sender, subscriber address) {
//	    if err := lifeSub.Subscribe(sender, subscriber); err != nil {
//	        panic(err)
//	    }
//	}
//
// In both cases, the caller's wrapper would typically construct an
// AllowanceSender capped at the per-period (or one-time) amount, defer
// its Close, and pass it as `sender`. If the consumer realm wants to
// restrict callers to EOAs (a policy choice unrelated to payment
// safety), the wrapper can additionally check
// runtime.PreviousRealm().IsUserCall().
package subscription
