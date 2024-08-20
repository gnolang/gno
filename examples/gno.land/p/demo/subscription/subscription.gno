package subscription

import (
	"std"
)

// Subscription interface defines standard methods that all subscription types must implement.
type Subscription interface {
	CheckAccess(std.Address) error
	ProcessPayment() error
	UpdateAmount(newAmount int64) error
}
