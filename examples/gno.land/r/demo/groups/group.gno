package groups

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/gor/checkpoints"
	"std"
	"strconv"
	"time"
)

type GroupID uint64

func (gid GroupID) String() string {
	return strconv.Itoa(int(gid))
}

type Group struct {
	id           GroupID
	url          string
	name         string
	lastMemberID MemberID
	members      avl.Tree
	upperAmount  int32 // upper amount of tokens a member hold, used to calculate tiers
	creator      std.Address
	createdAt    time.Time
}

func newGroup(url string, name string, creator std.Address) *Group {
	if !reName.MatchString(name) {
		panic("invalid name: " + name)
	}
	if gGroupsByName.Has(name) {
		panic("Group with such name already exists")
	}
	return &Group{
		id:        incGetGroupID(),
		url:       url,
		name:      name,
		creator:   creator,
		members:   avl.Tree{},
		createdAt: time.Now(),
	}
}

func (group *Group) newMember(id MemberID, address std.Address, weight int, metadata string) *Member {
	if group.members.Has(address.String()) {
		panic("this member for this group already exists")
	}
	member := &Member{
		id:                id,
		address:           address,
		weightCheckPoints: checkpoints.NewHistory(),
		metadata:          metadata,
		createdAt:         time.Now(),
	}
	member.weightCheckPoints.Push(uint64(weight))
	return member
}

func (group *Group) getTotalWeight() int32 {
	var weight int32
	group.members.Iterate("", "", func(key string, value interface{}) bool {
		member := value.(*Member)
		weight += int32(member.weightCheckPoints.LatestValue())
		println("weight accumulated to : ", weight)
		//weight += int32(member.weight)
		return false
	})
	return weight
}

// TODO: weight should be snapshot
func (group *Group) getMemberWeight(mid MemberID) int32 {
	var member *Member
	midkey := memberIDKey(mid)
	m, ok := group.members.Get(midkey)
	if !ok {
		panic("this member for this group not exists")
	}
	member = m.(*Member)
	return int32(member.weightCheckPoints.LatestValue())
}

func (group *Group) getMemberWeightAt(mid MemberID, blockNum int64) int32 {
	var member *Member
	midkey := memberIDKey(mid)
	m, ok := group.members.Get(midkey)
	if !ok {
		panic("this member for this group not exists")
	}
	member = m.(*Member)
	return int32(member.weightCheckPoints.GetAtBlock(blockNum))
}

// default at current height
func (group *Group) setWeightSnapshot(midkey string, weight uint64) {
	println("update weight snapshot: ", midkey, weight)
	var member *Member
	m, ok := group.members.Get(midkey)
	if !ok {
		panic("this member for this group not exists")
	}
	member = m.(*Member)

	member.weightCheckPoints.Push(weight)
	group.members.Set(midkey, member)
}

func segmentSearch(arr []int32, target int32) int {
	left, right := 0, len(arr)-1
	if target <= 0 || target > arr[right] {
		panic("should not happen")
	}

	for left <= right {
		mid := left + (right-left)/2
		if target > arr[mid] {
			if target <= arr[mid+1] {
				return mid + 2 // 1 for next range, 1 for 1 offset(start from 1), so plus 2
			}
			left = mid + 1 // shift
		} else {
			if mid == 0 {
				return 1
			}
			if target > arr[mid-1] {
				return mid + 1
			}
			right = mid - 1 // shift
		}
	}
	return -1
}

func tierCalculate(tierNum int32, upper int32, target int32) int {
	// TODO: some cache mechanism, so we don't need segememt calculating everytime
	// TODO: maybe some segmemt tree?
	var segments []int32
	var i int32
	for i = 0; i < tierNum; i++ {
		segments = append(segments, (i+1)*upper)
	}

	expansiveTarget := target * tierNum

	tier := segmentSearch(segments, expansiveTarget)

	return tier
}

// this set the weight by tier, which is calculated by tokens on hold, and a tier mechanism
func (group *Group) updateWeightSnapshotByTier(mid MemberID, amount int32, numTiers int32) {
	// upperLimit change, update all tiers
	if amount > group.upperAmount {
		group.upperAmount = amount
		var newWeight int32
		group.members.Iterate("", "", func(key string, value interface{}) bool {
			member := value.(*Member)
			if key == memberIDKey(mid) { // current payee
				newWeight = int32(tierCalculate(numTiers, group.upperAmount, amount))
				group.setWeightSnapshot(key, uint64(newWeight))
			} else {
				latest := member.weightCheckPoints.LatestValue()
				// reCalculate tier
				//newWeight = int32(calcWeightByTokenAmount(group.upperAmount, int32(latest), numTiers))
				newWeight = int32(tierCalculate(numTiers, group.upperAmount, int32(latest)))
				group.setWeightSnapshot(key, uint64(newWeight))
			}
			return false
		})
	}
	return
}

func (group *Group) HasPermission(addr std.Address, perm Permission) bool {
	if group.creator != addr {
		return false
	}
	return isValidPermission(perm)
}

func (group *Group) RenderGroup() string {
	str := "Group ID: " + groupIDKey(group.id) + "\n\n" +
		"Group Name: " + group.name + "\n\n" +
		"Group Creator: " + usernameOf(group.creator) + "\n\n" +
		"Group createdAt: " + group.createdAt.String() + "\n\n" +
		"Group Last MemberID: " + memberIDKey(group.lastMemberID) + "\n\n"

	str += "Group Members: \n\n"
	group.members.Iterate("", "", func(key string, value interface{}) bool {
		member := value.(*Member)
		str += member.getMemberStr()
		return false
	})
	return str
}

func (group *Group) deleteGroup() {
	gidkey := groupIDKey(group.id)
	_, gGroupsRemoved := gGroups.Remove(gidkey)
	if !gGroupsRemoved {
		panic("group does not exist with id " + group.id.String())
	}
	gGroupsByName.Remove(group.name)
}

// TODO: remove voting power
func (group *Group) deleteMember(mid MemberID) {
	gidkey := groupIDKey(group.id)
	if !gGroups.Has(gidkey) {
		panic("group does not exist with id " + group.id.String())
	}

	g := getGroup(group.id)
	midkey := memberIDKey(mid)
	g.members.Remove(midkey)
}
