package memberstore

import (
	"errors"

	"gno.land/p/nt/avl/v0"
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
	*avl.Tree // tier name -> address -> member
}

func NewMembersByTier() MembersByTier {
	return MembersByTier{Tree: avl.NewTree()}
}

func (mbt MembersByTier) DeleteAll() {
	mbt.Iterate("", "", func(tn string, msv interface{}) bool {
		mbt.Remove(tn)
		return false
	})
}

func (mbt MembersByTier) SetTier(tier string) error {
	if ok := mbt.Has(tier); ok {
		return errors.New("tier already exist: " + tier)
	}

	mbt.Set(tier, avl.NewTree())

	return nil
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
func (mbt MembersByTier) SetMember(tier string, addr address, member *Member) error {
	_, t := mbt.GetMember(addr)
	if t != "" {
		return &ErrMemberAlreadyExists{Tier: t}
	}

	if ok := mbt.Has(tier); !ok {
		return errors.New("tier does not exist: " + tier)
	}

	ms, _ := mbt.Get(tier)
	mst := ms.(*avl.Tree)

	mst.Set(string(addr), member)

	return nil
}

// GetMember iterate over all tiers to try to find a member by its address. The tier ID is also returned if the Member is found.
func (mbt MembersByTier) GetMember(addr address) (m *Member, t string) {
	mbt.Iterate("", "", func(tn string, msv interface{}) bool {
		mst, ok := msv.(*avl.Tree)
		if !ok {
			panic("MembersByTier values can only be avl.Tree")
		}

		mv, ok := mst.Get(string(addr))
		if !ok {
			return false
		}

		mm, ok := mv.(*Member)
		if !ok {
			panic("MembersByTier values can only be *Member")
		}

		m = mm
		t = tn

		return true
	})

	return
}

// RemoveMember removes a member from any tier
func (mbt MembersByTier) RemoveMember(addr address) (t string) {
	mbt.Iterate("", "", func(tn string, msv interface{}) bool {
		mst, ok := msv.(*avl.Tree)
		if !ok {
			panic("MembersByTier values can only be avl.Tree")
		}

		_, removed := mst.Remove(string(addr))
		if removed {
			t = tn
		}
		return removed
	})

	return
}

// GetTotalPower obtains the total voting power from all the specified tiers.
func (mbt MembersByTier) GetTotalPower() float64 {
	var out float64
	mbt.Iterate("", "", func(tn string, msv interface{}) bool {
		tier, ok := tiers.GetTier(tn)
		if !ok {
			// tier does not exists, so we cannot count power from this tier
			return false
		}

		out = out + (tier.PowerHandler(mbt, tiers) * float64(mbt.GetTierSize(tn)))

		return false
	})

	return out
}

type Tier struct {
	// BasePower defines the standard voting power for the members on this tier.
	BasePower float64

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
