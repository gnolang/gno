package subscription

// Subscription interface defines standard methods that all subscription types must implement.
type Subscription interface {
	HasValidSubscription(address) error
	Subscribe() error
	UpdateAmount(newAmount int64) error
}
