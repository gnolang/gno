package subscription

import "chain"

// Sender is the payment primitive subscription libraries accept. The
// caller hands a Sender capped at (or above) the subscription's price;
// the lib pulls exactly the per-period amount on a successful Subscribe
// call. Any cap-token type that exposes Send (e.g.
// gno.land/p/jaekwon/allowancesender/v0.AllowanceSender) satisfies it.
//
// Send must move real coins or panic. The implementation is
// expected to enforce its own caps, denom restrictions, and source
// authority — the subscription lib trusts whatever Send sees fit to
// move and only requires the payment to land at the lib's payeeAddr.
type Sender interface {
	Send(to address, amt chain.Coins)
}

// Subscription is the cross-implementation interface.
type Subscription interface {
	HasValidSubscription(address) error
	Subscribe(sender Sender, subscriber address) error
	UpdateAmount(newAmount int64) error
}
