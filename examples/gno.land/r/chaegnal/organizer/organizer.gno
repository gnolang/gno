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
	VoteYay = 1
	VoteNay = 0
)

var (
	ListGroups AllGroups

	NextGroupId = 1
)

type AllGroups struct {
	Groups *avl.Tree
}

type FriendGroup struct {
	ID    int
	Title string
	Owner *ownable.Ownable
	DAO   *dao.DAO
}

func init() {
	ListGroups.Groups = avl.NewTree()
}

// Create the new friendgroup with its associated DAO.
// Set the creator as the default admin
func CreateFriendGroup(title string) {
	crossing()
	caller := std.OriginCaller()

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

// Triggered when a user request to join a group
// Put the user in a waitlist until an admin let them join
// Verify if the group exist and if the user is already a member
func AskJoinFriendGroup(groupId int) {
	crossing()
	txSender := std.OriginCaller()

	raw, exists := ListGroups.Groups.Get(strconv.Itoa(groupId))
	if !exists {
		panic("group does not exist")
	}

	group := raw.(FriendGroup)

	_, ok := group.DAO.IsMember(txSender)
	if ok {
		panic("you are already part of this friend group")
	}

	group.DAO.PendingApproval(txSender)
	ListGroups.Groups.Set(strconv.Itoa(groupId), group)
}

// Wrapper to leave a group and DAO
func LeaveGroup(groupId string) {
	crossing()
	group := GetGroupFromId(groupId)

	if std.OriginCaller() == group.Owner.Owner() {
		panic("owner cannot leave the group")
	}

	group.DAO.LeaveDao(std.OriginCaller())
}

// Helpers

// Return the user's registered name or address
func ResolveAddressName(addr std.Address) string {
	user := users.ResolveAddress(addr)
	if user != nil {
		return user.Name()
	}
	return addr.String()
}

// Retrieve a group from a given ID
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
