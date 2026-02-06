package impl

import (
	"gno.land/p/nt/avl"
	"gno.land/r/gov/dao/v3/memberstore"
)

type MemberActivity struct {
	Available uint32
	Votes     uint32
}

// Global activity tracker: address -> MemberActivity
var memberActivityTree = avl.NewTree()

func incrementMemberActivity(addr address) {
	addrStr := addr.String()
	activity := MemberActivity{}
	if val, exists := memberActivityTree.Get(addrStr); exists {
		activity = val.(MemberActivity)
	}
	activity.Votes++
	memberActivityTree.Set(addrStr, activity)
}

func GetMemberActivity(addr address) MemberActivity {
	addrStr := addr.String()
	if val, exists := memberActivityTree.Get(addrStr); exists {
		return val.(MemberActivity)
	}
	return MemberActivity{}
}

func incrementAvailableVotesForAllMembers(cur realm) {
	tiers := memberstore.Get()

	tiers.Iterate("", "", func(tierName string, tierValue any) bool {
		memberTree, ok := tierValue.(*avl.Tree)
		if !ok {
			return false
		}

		memberTree.Iterate("", "", func(addrKey string, _ any) bool {
			activity := MemberActivity{}
			if val, exists := memberActivityTree.Get(addrKey); exists {
				activity = val.(MemberActivity)
			}
			activity.Available++
			memberActivityTree.Set(addrKey, activity)
			return false
		})
		return false
	})
}
