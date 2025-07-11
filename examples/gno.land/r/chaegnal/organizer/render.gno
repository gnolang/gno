package organizer

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/txlink"
)

const (
	pageSize = 5
)

// displayJoiningStatus displays the joining status for each group if user is logged in
func displayJoiningStatus(group FriendGroup, adr std.Address) string {
	var out string

	if adr.IsValid() {
		if group.DAO.Members().Has(adr) {
			out += md.Paragraph("You already are part of this group")
		} else if group.WaitList.Has(adr.String()) {
			out += md.Paragraph("You are awaiting approval for this group")
		} else {
			out += md.Link("Join Group", txlink.Call("AskJoinFriendGroup", "groupId", group.ID.String())) + "\n\n"
		}
	}

	return out
}

// displayAllExistingGroups displays all the existing group in a pager
// Display primitive informations about them and joining status if logged in
func displayAllExistingGroups(pageNumber, pageSize int, adr std.Address) string {
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

		author := NameOrAddress(group.Owner.Owner())

		out += md.H3("Group: "+group.Title+" - By: "+author) + "\n"
		out += md.Paragraph("Number of friends in the group: "+strconv.Itoa(group.DAO.Members().Size())) + "\n"

		out += displayJoiningStatus(group, adr)

		out += md.HorizontalRule() + "\n"
	}

	out += page.Picker("/organizer")

	return out
}

// displayGroupInfo display, in a table, the joined group and give the option to see their page or leave the group
// It does not allow pending member to view the group
func displayGroupInfo(group FriendGroup, table *mdtable.Table, hasJoined bool, adr std.Address) {
	author := NameOrAddress(group.Owner.Owner())
	var view string

	if hasJoined {
		view = md.Link("View", "/r/chaegnal/organizer/group?groupId="+group.ID.String()+"&address="+adr.String())

		if adr != group.Owner.Owner() {
			view += " | " + md.Link("Leave", txlink.Call("LeaveGroup", "groupId", group.ID.String()))
		}
	} else {
		view = "You cannot view group yet"
	}

	table.Append([]string{group.Title, author, strconv.Itoa(group.DAO.Members().Size()), view})
}

// displayJoinedGroups displays the group joined by the logged in user
func displayJoinedGroups(adr std.Address) string {
	var out string
	found := false

	out += md.H2("Your Joined Groups") + "\n"
	table := &mdtable.Table{
		Headers: []string{"Title", "Created By", "Members", "Inspect"},
	}

	ListGroups.Groups.Iterate("", "", func(key string, val any) bool {
		group := val.(FriendGroup)

		if group.DAO.Members().Has(adr) {
			displayGroupInfo(group, table, true, adr)
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

// displayPendingApproval displays the group that the user asked to join but is not yet approved
func displayPendingApproval(adr std.Address) string {
	var out string
	found := false

	table := &mdtable.Table{
		Headers: []string{"Title", "Created By", "Members", "Inspect"},
	}

	ListGroups.Groups.Iterate("", "", func(key string, val any) bool {
		group := val.(FriendGroup)

		if group.WaitList.Has(string(adr)) {
			displayGroupInfo(group, table, false, adr)
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
	adr := std.Address(addressStr)
	out += md.H2("Dashboard")
	if adr == "" {
		out += md.Paragraph("Please provide your address in the URL (?address=...) to get a personalized dashboard")
	} else {
		out += displayJoinedGroups(adr)
		out += md.H2("Your Pending Request")
		out += displayPendingApproval(adr)
	}

	out += md.HorizontalRule()
	out += md.H2("List of FriendGroups")
	out += displayAllExistingGroups(pageNumber, size, adr)

	return out
}
