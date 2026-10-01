package subscription

import (
	"gno.land/p/jaekwon/allowancesender/v0"
)

// Subscription is the cross-implementation interface. Payment is
// pulled from a caller-supplied *AllowanceSender; subscriber identity
// is decoupled from cap-token authority.
type Subscription interface {
	HasValidSubscription(address) error
	Subscribe(al *allowancesender.AllowanceSender, subscriber address) error
	UpdateAmount(newAmount int64) error
}
