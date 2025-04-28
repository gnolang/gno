package memberstore

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao"
)

var members MembersByTier
var Tiers TiersByName

const (
	T1 = "T1"
	T2 = "T2"
	T3 = "T3"
)

func init() {
	members = NewMembersByTier()

	Tiers = TiersByName{avl.NewTree()}
	Tiers.Set(T1, Tier{
		InvitationPoints: 3,
		MinSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 70
		},
		MaxSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 0
		},
		BasePower: 3,
		PowerHandler: func(membersByTier MembersByTier, tiersByName TiersByName) float64 {
			return 3
		},
	})

	Tiers.Set(T2, Tier{
		InvitationPoints: 2,
		MaxSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return membersByTier.GetTierSize(T1) * 2
		},
		MinSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return membersByTier.GetTierSize(T1) / 4
		},
		BasePower: 2,
		PowerHandler: func(membersByTier MembersByTier, tiersByName TiersByName) float64 {
			t1ms := float64(membersByTier.GetTierSize(T1))
			t1, _ := tiersByName.GetTier(T1)
			t2ms := float64(membersByTier.GetTierSize(T2))
			t2, _ := tiersByName.GetTier(T2)

			t1p := t1.BasePower * t1ms
			t2p := t2.BasePower * t2ms

			// capped to 2/3 of tier 1
			t1ptreshold := t1p * (2.0 / 3.0)
			if t2p > t1ptreshold {
				return t1ptreshold / t2ms
			}

			return t2.BasePower
		},
	})

	Tiers.Set(T3, Tier{
		InvitationPoints: 1,
		MaxSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 0
		},
		MinSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 0
		},
		BasePower: 1,
		PowerHandler: func(membersByTier MembersByTier, tiersByName TiersByName) float64 {
			t1ms := float64(membersByTier.GetTierSize(T1))
			t1, _ := tiersByName.GetTier(T1)
			t3ms := float64(membersByTier.GetTierSize(T3))
			t3, _ := tiersByName.GetTier(T3)

			t1p := t1.BasePower * t1ms
			t3p := t3.BasePower * t3ms

			//capped to 1/3 of tier 1
			t1ptreshold := t1p * (1.0 / 3.0)
			if t3p > t1ptreshold {
				return t1ptreshold / t3ms
			}

			return t3.BasePower
		},
	})

}

func Render(string) string {
	var sb strings.Builder

	sb.WriteString("# Memberstore Govdao v3:\n\n")

	members.Iterate("", "", func(tn string, ti interface{}) bool {
		tree, ok := ti.(*avl.Tree)
		if !ok {
			return false
		}

		tier, ok := Tiers.GetTier(tn)
		if !ok {
			return false
		}

		tp := (tier.PowerHandler(members, Tiers) * float64(members.GetTierSize(tn)))

		sb.WriteString(ufmt.Sprintf("- Tier %v contains %v members with power: %v\n", tn, tree.Size(), tp))

		return false
	})

	return sb.String()
}

// Get gets the Members store
func Get() MembersByTier {
	realm := std.CurrentRealm().PkgPath()
	if !dao.InAllowedDAOs(realm) {
		panic("this Realm is not allowed to get the Members data: " + realm)
	}

	return members
}
