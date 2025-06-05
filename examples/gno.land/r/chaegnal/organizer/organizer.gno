package organizer

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/p/nt/commondao"
	"gno.land/r/sys/users"
)

const (
	VoteYay = 1 // VoteYay represented yes votes
	VoteNay = 0 // VoteNay represented no votes

	memberRank = 0 // memberRank represents the basic right of user
	adminRank  = 1 // adminRank represents the rank of admin with heightened rights
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
	ID       int
	Title    string
	Owner    *ownable.Ownable
	DAO      *commondao.CommonDAO
	WaitList *avl.Tree
	Rank     *avl.Tree
}

func init() {
	ListGroups.Groups = avl.NewTree()
}

// CreateFriendGroup creates a new group with the chosen title
// The caller is set as the first admin and the owner
func CreateFriendGroup(title string) {
	crossing()
	caller := std.PreviousRealm().Address()

	newGroup := FriendGroup{
		ID:    NextGroupId,
		Title: title,
		Owner: ownable.NewWithAddress(caller),
		DAO: commondao.New(
			commondao.WithName(title),
			commondao.WithMember(caller),
		),
		WaitList: avl.NewTree(),
		Rank:     avl.NewTree(),
	}

	newGroup.DAO.Members().Add(caller)
	newGroup.Rank.Set(caller.String(), adminRank)

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
	memberSet := commondao.NewMemberSet(group.DAO.Members())

	if memberSet.Has(caller) {
		panic("you are already part of this friend group")
	}

	if group.WaitList.Has(caller.String()) {
		panic("already whitelisted")
	}

	group.WaitList.Set(caller.String(), true)
	ListGroups.Groups.Set(strconv.Itoa(groupId), group)
}

// LeaveGroup is a wrapper to leave a group and DAO
func LeaveGroup(groupId string) {
	crossing()
	group := GetGroupFromId(groupId)

	if std.PreviousRealm().Address() == group.Owner.Owner() {
		panic("owner cannot leave the group")
	}

	group.DAO.Members().Remove(std.PreviousRealm().Address())
}

// Helpers

// NameOrAddress returns the user's registered name if available
func NameOrAddress(addr std.Address) string {
	user := users.ResolveAddress(addr)
	if user != nil {
		name := user.Name()
		if name != "" {
			return name
		}
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
