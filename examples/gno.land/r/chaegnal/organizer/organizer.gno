package organizer

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/chaegnal/dao"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ownable"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/txlink"
	"gno.land/r/sys/users"
)

const (
	pageSize = 5

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

// Create the new friendgroup with its associated DAO.
// Set the creator as the default admin
func CreateFriendGroup(title string) {
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
	group := GetGroupFromId(groupId)

	if std.OriginCaller() == group.Owner.Owner() {
		panic("owner cannot leave the group")
	}

	group.DAO.LeaveDao(std.OriginCaller())
}

// Display the joining status for each group if user is logged in
func displayJoiningStatus(group FriendGroup, address std.Address) string {
	var out string

	if address != "" {
		_, ok := group.DAO.IsMember(address)
		if ok {
			out += md.Paragraph("You already are part of this group")
		} else if group.DAO.IsPending(address) {
			out += md.Paragraph("You are awaiting approval for this group")
		} else {
			out += md.Link("Join Group", txlink.Call("AskJoinFriendGroup", "groupId", strconv.Itoa(group.ID))) + "\n\n"
		}
	}

	return out
}

// Display all the existing group in a pager
// Display primitive informations about them
// Display joining status if logged in
func displayAllExistingGroups(pageNumber, pageSize int, address std.Address) string {
	var out string

	pager := pager.NewPager(ListGroups.Groups, pageSize, false)

	page := pager.GetPage(pageNumber)

	if page.TotalItems == 0 {
		out += "No friend groups available.\n"
		return out
	}

	for _, item := range page.Items {
		group, ok := item.Value.(FriendGroup)
		if !ok {
			continue
		}

		author := ResolveAddressName(group.Owner.Owner())

		out += md.H3("Group: "+group.Title+" - By: "+author) + "\n"
		out += md.Paragraph("Number of friends in the group: "+strconv.Itoa(group.DAO.Whitelist.Size())) + "\n"

		out += displayJoiningStatus(group, address)

		out += md.HorizontalRule() + "\n"
	}

	out += page.Picker("/organizer")

	return out
}

// Display, in a table, the joined group and give the option to see their page or leave the group
// Does not allow pending member to view the group
func displayGroupInfo(group FriendGroup, table *mdtable.Table, hasJoined bool, address std.Address) {
	author := ResolveAddressName(group.Owner.Owner())
	var view string

	if hasJoined {
		view = md.Link("View", "/r/chaegnal/organizer/group?groupId="+strconv.Itoa(group.ID)+"&address="+address.String())

		if address != group.Owner.Owner() {
			view += " | " + md.Link("Leave", txlink.Call("LeaveGroup", "groupId", strconv.Itoa(group.ID)))
		}
	} else {
		view = "You cannot view group yet"
	}

	table.Append([]string{group.Title, author, strconv.Itoa(group.DAO.Whitelist.Size()), view})
}

func displayJoinedGroups(address std.Address) string {
	var out string
	found := false

	out += md.H2("Your Joined Groups") + "\n"
	table := &mdtable.Table{
		Headers: []string{"Title", "Created By", "Members", "Inspect"},
	}

	ListGroups.Groups.Iterate("", "", func(key string, val any) bool {
		group := val.(FriendGroup)
		//
		_, ok := group.DAO.IsMember(address)
		if ok {
			displayGroupInfo(group, table, true, address)
			found = true
		}

		return false
	})

	if !found {
		out = md.Paragraph("You haven't joined any groups yet.")
	} else {
		out += table.String()
	}

	return out
}

// Display the group that the user asked to join but is not yet approved
func displayPendingApproval(address std.Address) string {
	var out string
	found := false

	table := &mdtable.Table{
		Headers: []string{"Title", "Created By", "Members", "Inspect"},
	}

	ListGroups.Groups.Iterate("", "", func(key string, val any) bool {
		group := val.(FriendGroup)
		//
		if group.DAO.IsPending(address) {
			displayGroupInfo(group, table, false, address)
			found = true
		}
		return false
	})

	if !found {
		out = md.Paragraph("No pending join request.")
	} else {
		out += table.String()
	}

	return out
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

func Render(rawURL string) string {
	var out string

	pageNumber, size, err := pager.NewPager(ListGroups.Groups, pageSize, false).ParseQuery(rawURL)
	if err != nil {
		pageNumber = 1
	}

	out += md.H1("Welcome to Friendly Organizer")
	out += md.Paragraph("Here you can join and create FriendGroups to organize your next hangout")

	out += md.HorizontalRule()
	out += md.H2(md.Link("Create New FriendGroup", txlink.Call("CreateFriendGroup", "title", "default")) + "\n\n")
	out += md.HorizontalRule()

	addressStr := parseQuery(rawURL)
	address := std.Address(addressStr)
	out += md.H2("Dashboard")
	if address == "" {
		out += md.Paragraph("Please provide your address in the URL (?address=...) to get a personalized dashboard")
	} else {
		out += displayJoinedGroups(address)
		out += md.H2("Your Pending Request")
		out += displayPendingApproval(address)
	}

	out += md.HorizontalRule()
	out += md.H2("List of FriendGroups")
	out += displayAllExistingGroups(pageNumber, size, address)

	return out
}
