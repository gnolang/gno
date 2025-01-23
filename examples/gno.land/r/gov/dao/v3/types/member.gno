package types

import (
	"std"

	"gno.land/p/demo/avl"
)

type Member struct {
	InvitationPoints int
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
func (mbt MembersByTier) SetMember(tier string, addr std.Address, member *Member) {
	// TODO: check if the member exists on another tier

	if ok := mbt.Has(tier); !ok {
		mbt.Set(tier, avl.NewTree())
	}

	ms, _ := mbt.Get(tier)
	mst := ms.(*avl.Tree)

	mst.Set(string(addr), member)
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

// GetTotalPower obtains the total voting power from all the specified tiers.
func (mbt MembersByTier) GetTotalPower(tbn TiersByName) float64 {
	var out float64
	mbt.Iterate("", "", func(tn string, msv interface{}) bool {
		tier, ok := tbn.GetTier(tn)
		if !ok {
			// tier does not exists, so we cannot count power from this tier
			return false
		}

		out = out + tier.PowerHandler(mbt, tbn)

		return false
	})

	return out
}
