package groups

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/teritori/havl"
	"gno.land/p/demo/teritori/ujson"
)

type GroupID uint64

func (gid GroupID) String() string {
	return strconv.Itoa(int(gid))
}

func (g *GroupID) FromJSON(ast *ujson.JSONASTNode) {
	if ast.ValueKind != ujson.JSONTokenKindNumber {
		panic("invalid type")
	}
	*g = GroupID(ujson.ParseUint64(ast.Value))
}

func (g GroupID) ToJSON() string {
	return ujson.FormatUint64(uint64(g))
}

type Group struct {
	id               GroupID
	url              string
	name             string
	lastMemberID     MemberID
	members          havl.Tree // id -> member
	membersByAddress havl.Tree // address -> member
	creator          std.Address
	createdAt        time.Time
	totalWeight      havl.Tree // "" -> int64
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
		createdAt: time.Now(),
	}
}

func (group *Group) newMember(id MemberID, address std.Address, weight int, metadata string) *Member {
	if group.members.Has(address.String(), 0) {
		panic("this member for this group already exists")
	}
	if group.membersByAddress.Has(address.String(), 0) {
		panic("this address is already in this group")
	}
	return &Member{
		id:        id,
		address:   address,
		weight:    weight,
		metadata:  metadata,
		createdAt: time.Now(),
	}
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
		"Group Creator: " + string(group.creator) + "\n\n" +
		"Group createdAt: " + group.createdAt.String() + "\n\n" +
		"Group Last MemberID: " + memberIDKey(group.lastMemberID) + "\n\n" +
		"Total Weight: " + strconv.Itoa(group.getTotalWeight(0)) + "\n\n"

	str += "Group Members: \n\n"
	group.members.Iterate("", "", 0, func(key string, value interface{}) bool {
		member := value.(*Member)
		str += member.getMemberStr()
		return false
	})

	return str
}

func (group *Group) GetMembers() *avl.Tree {
	snapshot, _ := group.members.GetSnapshot(0)
	return snapshot
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
	imember, ok := g.members.Get(midkey, 0)
	if !ok {
		return
	}
	member := imember.(*Member)
	g.members.Remove(midkey)
	g.membersByAddress.Remove(member.address.String())
	g.addToTotalWeight(-member.weight)
}

func (group *Group) addToTotalWeight(weight int) {
	current, ok := group.totalWeight.Get("", 0)
	if !ok {
		group.totalWeight.Set("", weight)
		return
	}
	group.totalWeight.Set("", current.(int)+weight)
}

func (group *Group) getTotalWeight(height int64) int {
	val, ok := group.totalWeight.Get("", height)
	if !ok {
		return 0
	}
	return val.(int)
}
