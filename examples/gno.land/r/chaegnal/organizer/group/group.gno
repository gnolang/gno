package group

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/txlink"
	"gno.land/p/sunspirit/md"

	"gno.land/p/chaegnal/dao"
	"gno.land/r/chaegnal/organizer"
)

// Parse the URL to retrieve the data behind the "=" sign
func parseQuery(url string) string {
	groupId := strings.SplitAfter(url, "?")
	foundId := strings.SplitAfter(groupId[1], "=")

	return foundId[1]
}

// Returns group associated with given ID
// Verify that the ID is valid and associated with a group
func getGroupFromId(groupId string) organizer.FriendGroup {
	convertedId, err := strconv.Atoi(groupId)
	if err != nil {
		panic("Invalid ID")
	}

	raw, found := organizer.ListGroups.Groups.Get(strconv.Itoa(convertedId))
	if !found {
		panic("Could not find group")
	}

	return raw.(organizer.FriendGroup)
}

// Retrieve the group ID from the URL
func getGroupFromUrl(url string) organizer.FriendGroup {
	groupId := parseQuery(url)
	if groupId == "" {
		panic("No groupId provided")
	}

	return getGroupFromId(groupId)
}

// Wrapper to accept new members into friendgroup and DAO
func AcceptAwaitingMember(groupId string, adr std.Address) {
	group := getGroupFromId(groupId)

	group.DAO.WhitelistingAwaitingApproval(organizer.AddressLoggedIn, adr)
}

// Wrapper to refuse new members into friendgroup and DAO
func RefuseAwaitingMember(groupId string, adr std.Address) {
	group := getGroupFromId(groupId)

	group.DAO.RefuseAwaitingMember(organizer.AddressLoggedIn, adr)
}

// Wrapper to remove members from group and DAO
func RemoveMember(groupId string, adr std.Address) {
	group := getGroupFromId(groupId)

	group.DAO.RemoveMember(organizer.AddressLoggedIn, adr)
}

// Wrapper to promote member to admin status
func MakeAdmin(groupId string, adr std.Address) {
	group := getGroupFromId(groupId)

	group.DAO.SetAdmin(organizer.AddressLoggedIn, adr)
}

// Wrapper to demote admin to member status
func RemoveAdminRank(groupId string, adr std.Address) {
	group := getGroupFromId(groupId)

	group.DAO.RemoveAdminRank(organizer.AddressLoggedIn, adr)
}

// Wrapper to create proposal
// Sanitizes the arguments given
func CreateProposals(groupId, title, desc, callback string, callbackArgsStr string) {
	group := getGroupFromId(groupId)

	callbackArgsStr = strings.Trim(callbackArgsStr, "\"")
	callbackArgs := strings.Split(callbackArgsStr, ",")

	group.DAO.CreateProposals(title, desc, callback, callbackArgs, organizer.AddressLoggedIn)
}

// Allows the owner to delete the friendgroup
func DeleteFriendGroup(groupId string) {
	group := getGroupFromId(groupId)

	if group.Owner.Owner() != organizer.AddressLoggedIn {
		panic("Only the owner can delete the group.")
	}

	organizer.ListGroups.Groups.Remove(strconv.Itoa(group.ID))
}

func SwitchOwner(groupId string, adr std.Address) {
	group := getGroupFromId(groupId)

	if group.Owner.Owner() != organizer.AddressLoggedIn {
		panic("Only the owner give ownership away.")
	}

	ok := group.DAO.IsAdmin(adr)
	if !ok {
		panic("The new owner must be an admin")
	}

	group.Owner = ownable.NewWithAddress(adr)
	organizer.ListGroups.Groups.Set(strconv.Itoa(group.ID), group)
}

// Wrapper to vote on proposal
// Verify if the vote ID exist and if the vote's choice exists
func VoteOnProposal(groupId string, voteIdStr string, voteStr string) {
	group := getGroupFromId(groupId)

	voteId, err := strconv.Atoi(voteIdStr)
	if err != nil {
		panic("Invalid vote ID.")
	}

	vote, err := strconv.Atoi(voteStr)
	if err != nil {
		panic("Invalid vote choice.")
	}

	if vote != organizer.VoteNay && vote != organizer.VoteYay {
		panic("Vote should be Yay (1) or Nay (0).")
	}

	group.DAO.VoteProposal(voteId, vote, organizer.AddressLoggedIn)
}

// Display in a table the users awaiting approval into the group
// Let admin directly accept or refuse members
func displayPendingApproval(group organizer.FriendGroup, awaiting *avl.Tree) string {
	table := mdtable.Table{
		Headers: []string{"Address", "Inspect"},
	}

	awaiting.Iterate("", "", func(indexStr string, val any) bool {
		adr := std.Address(indexStr)

		table.Append([]string{adr.String(),
			md.Link("Approve", txlink.Call("AcceptAwaitingMember", "groupId", strconv.Itoa(group.ID), "adr", adr.String())) + " | " +
				md.Link("Reject", txlink.Call("RefuseAwaitingMember", "groupId", strconv.Itoa(group.ID), "adr", adr.String())),
		})
		return false
	})

	return table.String()
}

// Display the group current members
// Let the admin promote / demote / remove member
func displayMembers(group organizer.FriendGroup, whitelist *avl.Tree) string {
	table := mdtable.Table{
		Headers: []string{"Address", "Action"},
	}

	whitelist.Iterate("", "", func(indexStr string, val any) bool {
		var actionAdmin string

		adr := std.Address(indexStr)
		if adr == organizer.AddressLoggedIn {
			return false
		}

		if group.DAO.IsAdmin(adr) {
			actionAdmin = md.Link("Remove Admin", txlink.Call("RemoveAdminRank", "groupId", strconv.Itoa(group.ID), "adr", adr.String()))
		} else {
			actionAdmin = md.Link("Make Admin", txlink.Call("MakeAdmin", "groupId", strconv.Itoa(group.ID), "adr", adr.String()))
		}

		table.Append([]string{adr.String(),
			actionAdmin + " | " +
				md.Link("Kick out", txlink.Call("RemoveMember", "groupId", strconv.Itoa(group.ID), "adr", adr.String())),
		})

		return false
	})

	return table.String()
}

// Display admin board with Waitlist and Whiteliste
// Let admins create new proposals
func displayAdminBoard(group organizer.FriendGroup) string {
	var out string

	out += md.H2("Admin Board") + "\n"

	out += md.Link("Create Proposal", txlink.Call("CreateProposals",
		"groupId", strconv.Itoa(group.ID), "title", "desc", "callback", "callbackArgs")) + "\n"

	out += md.H3("Approval Waitlist") + "\n"

	awaiting := group.DAO.GetAwaiting()
	if awaiting.Size() > 0 {
		out += displayPendingApproval(group, awaiting)
	} else {
		out += md.Paragraph("No member waiting for approval") + "\n"
	}
	out += md.HorizontalRule() + "\n"

	out += md.H3("Members") + "\n"

	whitelist := group.DAO.GetWhitelisted()
	if whitelist.Size() > 1 {
		out += displayMembers(group, whitelist)
	} else {
		out += md.Paragraph("You're the only member for now.") + "\n"
	}
	out += md.HorizontalRule() + "\n"

	return out
}

func displayProposals(group organizer.FriendGroup, activeOnly bool) string {
	proposals := group.DAO.GetProposals()

	table := mdtable.Table{
		Headers: []string{"Title", "Description", "Status", "Votes", "Actions", "Result"},
	}

	proposals.Iterate("", "", func(index string, val any) bool {
		proposal := val.(dao.Proposal)

		if activeOnly != proposal.Active {
			return false
		}

		voteYayRaw, _ := proposal.Votes.Get(strconv.Itoa(organizer.VoteYay))
		voteNayRaw, _ := proposal.Votes.Get(strconv.Itoa(organizer.VoteNay))

		yay := 0
		nay := 0
		if voteYayRaw != nil {
			yay = voteYayRaw.(int)
		}
		if voteNayRaw != nil {
			nay = voteNayRaw.(int)
		}

		status := "Closed"
		if proposal.Active {
			status = "Active"
		}

		result := "-"
		if !proposal.Active {
			result = proposal.Winner
		}

		voteActions := "-"
		if proposal.Active {
			voteActions =
				md.Link("Yay", txlink.Call("VoteOnProposal",
					"groupId", strconv.Itoa(group.ID),
					"index", index,
					"vote", strconv.Itoa(organizer.VoteYay))) + " | " +
					md.Link("Nay", txlink.Call("VoteOnProposal",
						"groupId", strconv.Itoa(group.ID),
						"index", index,
						"vote", strconv.Itoa(organizer.VoteNay)))
		}

		table.Append([]string{
			proposal.Title,
			proposal.Description,
			status,
			"Yay: " + strconv.Itoa(yay) + " / Nay: " + strconv.Itoa(nay),
			voteActions,
			result,
		})

		return false
	})

	return table.String()
}

func Render(url string) string {
	var out string

	group := getGroupFromUrl(url)

	out += md.H1("Group: " + group.Title)
	out += md.H2("Created by: " + organizer.ResolveAddressName(group.Owner.Owner()))
	out += md.Paragraph("Number of friends: "+strconv.Itoa(group.DAO.Whitelist.Size())) + "\n"
	out += md.HorizontalRule()

	if group.DAO.IsAdmin(organizer.AddressLoggedIn) {
		out += displayAdminBoard(group)
	}

	out += md.H2("Active DAOs")
	out += displayProposals(group, true)
	out += md.HorizontalRule()

	out += md.H2("Closed DAOs")
	out += displayProposals(group, false)
	out += md.HorizontalRule()

	out += md.Paragraph(md.Link("Go back", "/r/chaegnal/organizer"))

	if group.Owner.Owner() == organizer.AddressLoggedIn {
		out += md.Paragraph(" | " + md.Link("Delete Group", txlink.Call("DeleteFriendGroup", "groupId", strconv.Itoa(group.ID))))
	}
	return out
}
