package groups

import (
	"gno.land/p/demo/avl"
	"std"
	"strconv"
	"time"
)

type GroupID uint64
type MemberID uint64

func (gid GroupID) String() string {
	return strconv.Itoa(int(gid))
}

func (mid MemberID) String() string {
	return strconv.Itoa(int(mid))
}

type Member struct {
	id        MemberID
	address   string
	weight    int
	metadata  string
	createdAt time.Time
}

type Group struct {
	id           GroupID
	url          string
	name         string
	lastMemberID MemberID
	members      *avl.MutTree
	creator      std.Address
	createdAt    time.Time
}

func newGroup(id GroupID, url string, name string, creator std.Address) *Group {
	if !reName.MatchString(name) {
		panic("invalid name: " + name)
	}
	exists := gGroupsByName.Has(name)
	if exists {
		panic("Group with such name already exists")
	}
	return &Group{
		id:        id,
		url:       url,
		name:      name,
		creator:   creator,
		members:   avl.NewMutTree(),
		createdAt: time.Now(),
	}
}

func newMember(id MemberID, gid GroupID, address string, weight int, metadata string) *Member {

	group := getGroup(gid)
	exists := group.members.Has(address)
	if exists {
		panic("this member for this group already exists")
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
	if group.creator == addr {
		switch perm {
		case EditPermission:
			return true
		case DeletePermission:
			return true
		default:
			return false
		}
	}
	return false
}

func (group *Group) RenderGroup() string {

	str := "Group ID: " + groupIDKey(group.id) + "\n\n" +
		"Group Name: " + group.name + "\n\n" +
		"Group Creator: " + usernameOf(group.creator) + "\n\n" +
		"Group createdAt: " + group.createdAt.String() + "\n\n" +
		"Group Last MemberID: " + memberIDKey(group.lastMemberID) + "\n\n"

	str += "Group Members: \n\n"

	group.members.Iterate("", "", func(n *avl.Tree) bool {
		member := n.Value().(*Member)
		str += "\t\t\t[" + memberIDKey(member.id) + ", " + member.address + ", " + strconv.Itoa(member.weight) + ", " + member.metadata + ", " + member.createdAt.String() + "],\n\n"
		return false
	})

	return str
}

func (group *Group) deleteGroup(gid GroupID) {

	gidkey := groupIDKey(gid)
	_, removed := gGroups.Remove(gidkey)
	if !removed {
		panic("group does not exist with id " + gid.String())
	}

}

func (group *Group) deleteMember(gid GroupID, mid MemberID) {

	gidkey := groupIDKey(gid)
	exists := gGroups.Has(gidkey)
	if !exists {
		panic("group does not exist with id " + gid.String())
	}

	g := getGroup(gid)
	midkey := memberIDKey(mid)
	_, removed := g.members.Remove(midkey)
	if !removed {
		panic("member does not exist with id " + mid.String())
	}

}
