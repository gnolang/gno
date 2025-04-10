package subscription

import "chain"

// Subscription interface defines standard methods that all subscription types must implement.
type Subscription interface {
	HasValidSubscription(chain.Address) error
	Subscribe() error
	UpdateAmount(newAmount int64) error
}
