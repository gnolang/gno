package organizer

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/chaegnal/dao"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/r/sys/users"
)

const (
	VoteYay = 1 // VoteYay represented yes votes
	VoteNay = 0 // VoteNay represented no votes
)

var (
	ListGroups AllGroups // ListGroups has all the created friend group

	NextGroupId = 1 // NextGroupId tracks the next available ID
)

// AllGroups wrap a avl.tree of created groups
type AllGroups struct {
	Groups *avl.Tree
}

// FriendGroup is a struc for created groups with their ID, Title, Owner and DAO
type FriendGroup struct {
	ID    int
	Title string
	Owner *ownable.Ownable
	DAO   *dao.DAO
}

func init() {
	ListGroups.Groups = avl.NewTree()
}

// CreateFriendGroup creates a new group with the chosen title
// The caller is set as the first admin and the owner
func CreateFriendGroup(title string) {
	crossing()
	caller := std.PreviousRealm().Address()

	newDao := dao.NewDAO(caller)
	newGroup := FriendGroup{
		ID:    NextGroupId,
		Title: title,
		Owner: ownable.NewWithAddress(caller),
		DAO:   newDao,
	}

	NextGroupId++
	ListGroups.Groups.Set(strconv.Itoa(newGroup.ID), newGroup)
}

// JoinFriendGroup puts the caller in a waitlist for the specified group
// The caller is in the waitlist until an admin of the group approves
func AskJoinFriendGroup(groupId int) {
	crossing()
	caller := std.PreviousRealm().Address()

	raw, exists := ListGroups.Groups.Get(strconv.Itoa(groupId))
	if !exists {
		panic("group does not exist")
	}

	group := raw.(FriendGroup)

	_, ok := group.DAO.IsMember(caller)
	if ok {
		panic("you are already part of this friend group")
	}

	group.DAO.PendingApproval(caller)
	ListGroups.Groups.Set(strconv.Itoa(groupId), group)
}

// LeaveGroup is a wrapper to leave a group and DAO
func LeaveGroup(groupId string) {
	crossing()
	group := GetGroupFromId(groupId)

	if std.PreviousRealm().Address() == group.Owner.Owner() {
		panic("owner cannot leave the group")
	}

	group.DAO.LeaveDao(std.PreviousRealm().Address())
}

// Helpers

// ResolveAddressName the user's registered name or address
func ResolveAddressName(addr std.Address) string {
	user := users.ResolveAddress(addr)
	if user != nil {
		return user.Name()
	}
	return addr.String()
}

// GetGroupFromId retrieves a group from a given ID
func GetGroupFromId(groupId string) FriendGroup {
	convertedId, err := strconv.Atoi(groupId)
	if err != nil {
		panic("invalid id")
	}

	raw, found := ListGroups.Groups.Get(strconv.Itoa(convertedId))
	if !found {
		panic("could not find group")
	}

	return raw.(FriendGroup)
}

func parseQuery(url string) string {
	parts := strings.SplitAfter(url, "?")
	if len(parts) < 2 {
		return ""
	}

	query := parts[1]
	keyValue := strings.SplitAfter(query, "=")
	if len(keyValue) < 2 {
		return ""
	}

	return keyValue[1]
}
