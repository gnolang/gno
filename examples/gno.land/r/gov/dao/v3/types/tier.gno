package types

import (
	"gno.land/p/demo/avl"
)

type Tier struct {
	// BasePower defines the standard voting power for the members on this tier.
	BasePower float64

	// Payment defines what will be the amount being paid to Members on different tiers. The payment itself will be good, but not exceed 90th percentile of senior software architect roles in the highest paid city globally.
	Payment int // TODO: implement payment.

	// InvitationPoints defines how many invitation points users on that tier will receive.
	InvitationPoints int

	// MaxSize calculates the max amount of members expected to be on this tier.
	MaxSize func(membersByTier MembersByTier, tiersByName TiersByName) int

	// MinSize calculates the min amount of members expected to be on this tier.
	MinSize func(membersByTier MembersByTier, tiersByName TiersByName) int

	// PowerHandler calculates what is the final power of this tier after taking into account Members by other tiers.
	PowerHandler func(membersByTier MembersByTier, tiersByName TiersByName) float64
}

// TiersByName contains all tier objects indexed by its name.
type TiersByName struct {
	*avl.Tree // *avl.Tree[string]Tier
}

// GetTier obtains a Tier struct by its name. It returns false if the Tier is not found.
func (tbn TiersByName) GetTier(tn string) (Tier, bool) {
	val, ok := tbn.Get(tn)
	if !ok {
		return Tier{}, false
	}

	t, ok := val.(Tier)
	if !ok {
		panic("TiersByName must contains only Tier types")
	}

	return t, true
}