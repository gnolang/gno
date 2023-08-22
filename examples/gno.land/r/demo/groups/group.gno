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
	return &Member{
		id:                id,
		address:           address,
		weight:            weight,
		weightCheckPoints: checkpoints.NewHistory(),
		metadata:          metadata,
		createdAt:         time.Now(),
	}
}

func (group *Group) getTotalWeight() int32 {
	var weight int32
	group.members.Iterate("", "", func(key string, value interface{}) bool {
		member := value.(*Member)
		weight += int32(member.weight)
		return false
	})
	return weight
}

// TODO: weight should be snapshot
func (group *Group) getMemberWeight(mid MemberID) int {
	var member *Member
	midkey := memberIDKey(mid)
	m, ok := group.members.Get(midkey)
	if !ok {
		panic("this member for this group not exists")
	}
	member = m.(*Member)
	return member.weight
}

func (group *Group) setMemberWeight(mid MemberID, weight int) {
	var member *Member
	midkey := memberIDKey(mid)
	m, ok := group.members.Get(midkey)
	if !ok {
		panic("this member for this group not exists")
	}
	member = m.(*Member)

	member.weight = weight

	group.members.Set(midkey, member)
}

// default at current height
func (group *Group) updateWeightSnapshot(midkey string, weight uint64) {
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

func calcWeightByTokenAmount(upperLimit, amount int32, numTiers int32) int32 {
	println("calc weight, upperLimit, amount ,numTiers: ", upperLimit, amount, numTiers)
	rangeSize := upperLimit / numTiers
	println("rangeSize is: ", rangeSize)
	tier := (amount-1)/rangeSize + 1
	return tier
}

// this set the weight by tier, which is calculated by tokens on hold, and a tier mechanism
func (group *Group) updateWeightSnapshotByTier(mid MemberID, amount int32, numTiers int32) {
	println("update weight by tier: ", mid.String(), amount)
	// upperLimit change, update all tiers
	if amount > group.upperAmount {
		group.upperAmount = amount
		var newWeight int32
		group.members.Iterate("", "", func(key string, value interface{}) bool {
			member := value.(*Member)
			if key == memberIDKey(mid) { // current payee
				newWeight = int32(calcWeightByTokenAmount(group.upperAmount, amount, numTiers))
				println("new weight is: ", newWeight)
				group.updateWeightSnapshot(key, uint64(newWeight))
			} else {
				latest := member.weightCheckPoints.LatestValue()
				println("member: ", key)
				println("latest is: ", latest)
				// reCalculate tier
				newWeight = int32(calcWeightByTokenAmount(group.upperAmount, int32(latest), numTiers))
				println("new weight is: ", newWeight)
				group.updateWeightSnapshot(key, uint64(newWeight))
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

func (group *Group) deleteMember(mid MemberID) {
	gidkey := groupIDKey(group.id)
	if !gGroups.Has(gidkey) {
		panic("group does not exist with id " + group.id.String())
	}

	g := getGroup(group.id)
	midkey := memberIDKey(mid)
	g.members.Remove(midkey)
}
