// Package subscription provides a flexible system for managing both recurring and
// lifetime subscriptions in Gno applications. It enables developers to handle
// payment-based access control for services or products. The library supports
// both subscriptions requiring periodic payments (recurring) and one-time payments
// (lifetime). Subscriptions are tracked using an AVL tree for efficient management
// of subscription statuses.
//
// Usage:
//
// Import the required sub-packages (`recurring` and/or `lifetime`) to manage specific
// subscription types. The methods provided allow users to subscribe, check subscription
// status, and manage payments.
//
// Recurring Subscription:
//
// Recurring subscriptions require periodic payments to maintain access.
// Users pay to extend their access for a specific duration.
//
// Example:
//
//	// Create a recurring subscription requiring 100 ugnot every 30 days
//	recSub := recurring.NewRecurringSubscription(time.Hour * 24 * 30, 100)
//
//	// Process payment for the recurring subscription
//	recSub.Subscribe()
//
//	// Gift a recurring subscription to another user
//	recSub.GiftSubscription(recipientAddress)
//
//	// Check if a user has a valid subscription
//	recSub.HasValidSubscription(addr)
//
//	// Get the expiration date of the subscription
//	recSub.GetExpiration(caller)
//
//	// Update the subscription amount to 200 ugnot
//	recSub.UpdateAmount(200)
//
//	// Get the current subscription amount
//	recSub.GetAmount()
//
// Lifetime Subscription:
//
// Lifetime subscriptions require a one-time payment for permanent access.
// Once paid, users have indefinite access without further payments.
//
// Example:
//
//	// Create a lifetime subscription costing 500 ugnot
//	lifeSub := lifetime.NewLifetimeSubscription(500)
//
//	// Process payment for lifetime access
//	lifeSub.Subscribe()
//
//	// Gift a lifetime subscription to another user
//	lifeSub.GiftSubscription(recipientAddress)
//
//	// Check if a user has a valid subscription
//	lifeSub.HasValidSubscription(addr)
//
//	// Update the lifetime subscription amount to 1000 ugnot
//	lifeSub.UpdateAmount(1000)
//
//	// Get the current lifetime subscription amount
//	lifeSub.GetAmount()
package subscription
