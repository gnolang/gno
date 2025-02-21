package memberstore

import (
	"std"

	"gno.land/p/demo/avl"
)

type ErrMemberAlreadyExists struct {
	Tier string
}

func (e *ErrMemberAlreadyExists) Error() string {
	return "member already exists on tier " + e.Tier
}

type Member struct {
	InvitationPoints int
}

func (m *Member) RemoveInvitationPoint() {
	if m.InvitationPoints <= 0 {
		panic("not enough invitation points")
	}

	m.InvitationPoints = m.InvitationPoints - 1
}

// MembersByTier contains all `Member`s indexed by their Address.
type MembersByTier struct {
	*avl.Tree // *avl.Tree[string]*avl.Tree[std.Address]*Member
}

func NewMembersByTier() MembersByTier {
	return MembersByTier{avl.NewTree()}
}

// GetTierSize tries to get how many members are on the specified tier. If the tier does not exists, it returns 0.
func (mbt MembersByTier) GetTierSize(tn string) int {
	tv, ok := mbt.Get(tn)
	if !ok {
		return 0
	}

	tree, ok := tv.(*avl.Tree)
	if !ok {
		return 0
	}

	return tree.Size()
}

// SetMember adds a new member to the specified tier. The tier index is created on the fly if it does not exists.
func (mbt MembersByTier) SetMember(tier string, addr std.Address, member *Member) error {
	_, t := mbt.GetMember(addr)
	if t != "" {
		return &ErrMemberAlreadyExists{Tier: t}
	}

	if ok := mbt.Has(tier); !ok {
		mbt.Set(tier, avl.NewTree())
	}

	ms, _ := mbt.Get(tier)
	mst := ms.(*avl.Tree)

	mst.Set(string(addr), member)

	return nil
}

// GetMember iterate over all tiers to try to find a member by its address. The tier ID is also returned if the Member is found.
func (mbt MembersByTier) GetMember(addr std.Address) (m *Member, t string) {
	mbt.Iterate("", "", func(tn string, msv interface{}) bool {
		mst, ok := msv.(*avl.Tree)
		if !ok {
			panic("MembersByTier values can only be avl.Tree")
		}

		return mst.Iterate("", "", func(ma string, msv interface{}) bool {
			if ma != string(addr) {
				return false
			}

			mst, ok := msv.(*Member)
			if !ok {
				panic("MembersByTier values can only be *Member")
			}

			m = mst
			t = tn

			return true
		})
	})

	return
}

// RemoveMember removes a member from any tier
func (mbt MembersByTier) RemoveMember(addr std.Address) (t string) {
	mbt.Iterate("", "", func(tn string, msv interface{}) bool {
		mst, ok := msv.(*avl.Tree)
		if !ok {
			panic("MembersByTier values can only be avl.Tree")
		}

		_, removed := mst.Remove(string(addr))
		t = tn
		return removed
	})

	return
}

// GetTotalPower obtains the total voting power from all the specified tiers.
func (mbt MembersByTier) GetTotalPower() float64 {
	var out float64
	mbt.Iterate("", "", func(tn string, msv interface{}) bool {
		tier, ok := Tiers.GetTier(tn)
		if !ok {
			// tier does not exists, so we cannot count power from this tier
			return false
		}

		out = out + (tier.PowerHandler(mbt, Tiers) * float64(mbt.GetTierSize(tn)))

		return false
	})

	return out
}

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
