package groups

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
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

func (group *Group) newMember(id MemberID, address_XXX std.Address, weight int, metadata string) *Member {
	if group.members.Has(address_XXX.String()) {
		panic("this member for this group already exists")
	}
	return &Member{
		id:          id,
		address_XXX: address_XXX,
		weight:      weight,
		metadata:    metadata,
		createdAt:   time.Now(),
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
		"Group Creator: " + usernameOf(group.creator) + "\n\n" +
		"Group createdAt: " + group.createdAt.String() + "\n\n" +
		"Group Last MemberID: " + memberIDKey(group.lastMemberID) + "\n\n"

	str += "Group Members: \n\n"
	group.members.Iterate("", "", func(key string, value any) bool {
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
