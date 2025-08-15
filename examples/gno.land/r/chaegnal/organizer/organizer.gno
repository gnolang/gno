package organizer

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/seqid"
	"gno.land/p/nt/commondao"
	"gno.land/r/sys/users"
)

const (
	VoteYay = 1 // VoteYay represented yes votes
	VoteNay = 0 // VoteNay represented no votes
)

const (
	MemberRank = iota // memberRank represents the basic right of user
	AdminRank         // adminRank represents the rank of admin with heightened rights
)

var (
	ListGroups AllGroups // ListGroups has all the created friend group
)

// AllGroups wrap a avl.tree of created groups
type AllGroups struct {
	Groups      *avl.Tree
	NextGroupId seqid.ID
}

// FriendGroup is a struct for created groups with their ID, Title, Owner and DAO
type FriendGroup struct {
	nextHangoutID seqid.ID
	ID            seqid.ID
	Title         string
	Owner         *ownable.Ownable
	DAO           *commondao.CommonDAO
	WaitList      *avl.Tree
	Rank          *avl.Tree
	Hangout       *avl.Tree
}

func init() {
	ListGroups.Groups = avl.NewTree()
}

// CreateFriendGroup creates a new group with the chosen title
// The caller is set as the first admin and the owner
func CreateFriendGroup(cur realm, title string) {
	caller := std.PreviousRealm().Address()

	id := ListGroups.NextGroupId.Next()

	newGroup := FriendGroup{
		ID:    id,
		Title: title,
		Owner: ownable.NewWithAddress(caller),
		DAO: commondao.New(
			commondao.WithName(title),
			commondao.WithMember(caller),
		),
		WaitList: avl.NewTree(),
		Rank:     avl.NewTree(),
		Hangout:  avl.NewTree(),
	}

	newGroup.Rank.Set(caller.String(), AdminRank)

	ListGroups.Groups.Set(id.String(), newGroup)
}

// AskJoinFriendGroup puts the caller in a waitlist for the specified group
// The caller is in the waitlist until an admin of the group approves
func AskJoinFriendGroup(cur realm, groupId int) {
	caller := std.PreviousRealm().Address()

	raw, exists := ListGroups.Groups.Get(seqid.ID(groupId).String())
	if !exists {
		panic("group does not exist")
	}

	group := raw.(FriendGroup)

	if group.DAO.Members().Has(caller) {
		panic("you are already part of this friend group")
	}

	if group.WaitList.Has(caller.String()) {
		panic("already whitelisted")
	}

	group.WaitList.Set(caller.String(), true)
	ListGroups.Groups.Set(seqid.ID(groupId).String(), group)
}

// LeaveGroup is a wrapper to leave a group and DAO
func LeaveGroup(cur realm, groupId string) {
	group := GetGroupFromId(groupId)

	if std.PreviousRealm().Address() == group.Owner.Owner() {
		panic("owner cannot leave the group")
	}

	group.DAO.Members().Remove(std.PreviousRealm().Address())
}

// Helpers

// NameOrAddress returns the user's registered name if available
func NameOrAddress(adr std.Address) string {
	user := users.ResolveAddress(adr)
	if user != nil {
		name := user.Name()
		if name != "" {
			return name
		}
	}
	return adr.String()
}

// GetGroupFromId retrieves a group from a given ID
func GetGroupFromId(groupId string) FriendGroup {
	convertedId, err := strconv.Atoi(groupId)
	if err != nil {
		panic("invalid id")
	}

	raw, found := ListGroups.Groups.Get(seqid.ID(convertedId).String())
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
