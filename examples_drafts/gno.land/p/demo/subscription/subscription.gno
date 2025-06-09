package subscription

import (
	"std"
)

// Subscription interface defines standard methods that all subscription types must implement.
type Subscription interface {
	HasValidSubscription(std.Address) error
	Subscribe() error
	UpdateAmount(newAmount int64) error
}
