package group

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/ownable"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/txlink"
	"gno.land/p/sunspirit/md"

	"gno.land/p/chaegnal/dao"
	"gno.land/r/chaegnal/organizer"
)

// Parse the URL to retrieve the data behind the "=" sign
func parseQueryGroupId(url string) string {
	query := strings.SplitAfter(url, "?")[1]
	params := strings.Split(query, "&")

	for _, param := range params {
		if strings.HasPrefix(param, "groupId=") {
			return strings.SplitAfter(param, "=")[1]
		}
	}
	return ""
}

func parseQueryAddress(url string) string {
	query := strings.SplitAfter(url, "?")[1]
	params := strings.Split(query, "&")

	for _, param := range params {
		if strings.HasPrefix(param, "address=") {
			return strings.SplitAfter(param, "=")[1]
		}
	}
	return ""
}

// Retrieve the group ID from the URL
func getGroupFromUrl(url string) organizer.FriendGroup {
	groupId := parseQueryGroupId(url)
	if groupId == "" {
		panic("no groupid provided")
	}

	return organizer.GetGroupFromId(groupId)
}

// Wrapper to accept new members into friendgroup and DAO
func AcceptAwaitingMember(groupId, logged string, adr std.Address) {
	group := organizer.GetGroupFromId(groupId)

	group.DAO.WhitelistUser(std.Address(logged), adr)
}

// Wrapper to refuse new members into friendgroup and DAO
func RefuseAwaitingMember(groupId, logged string, adr std.Address) {
	group := organizer.GetGroupFromId(groupId)

	group.DAO.RefuseAwaitingMember(std.Address(logged), adr)
}

// Wrapper to remove members from group and DAO
func RemoveMember(groupId, logged string, adr std.Address) {
	group := organizer.GetGroupFromId(groupId)

	group.DAO.RemoveMember(std.Address(logged), adr)
}

// Wrapper to promote member to admin status
func MakeAdmin(groupId, logged string, adr std.Address) {
	group := organizer.GetGroupFromId(groupId)

	group.DAO.SetAdmin(std.Address(logged), adr)
}

// Wrapper to demote admin to member status
func RemoveAdminRank(groupId, logged string, adr std.Address) {
	group := organizer.GetGroupFromId(groupId)

	if adr == group.Owner.Owner() {
		panic("you cannot undo admin rights of the owner")
	}
	group.DAO.RemoveAdminRank(std.Address(logged), adr)
}

// Wrapper to create proposal
// Sanitizes the arguments given
func CreateProposals(groupId, title, desc, callback, callbackArgsStr, logged string) {
	group := organizer.GetGroupFromId(groupId)

	callbackArgsStr = strings.Trim(callbackArgsStr, "\"")
	callbackArgs := strings.Split(callbackArgsStr, ",")

	group.DAO.CreateProposals(title, desc, callback, callbackArgs, std.Address(logged))
}

// Wrapper for admin to close proposal forcefully
func AdminCloseProposal(groupId, voteIdStr, logged string) {
	group := organizer.GetGroupFromId(groupId)

	voteId, err := strconv.Atoi(voteIdStr)
	if err != nil {
		panic("invalid vote id")
	}

	group.DAO.AdminCloseProposal(voteId, std.Address(logged))
}

// Allows the owner to delete the friendgroup
func DeleteFriendGroup(groupId, logged string) {
	group := organizer.GetGroupFromId(groupId)

	if group.Owner.Owner() != std.Address(logged) {
		panic("only the owner can delete the group")
	}

	organizer.ListGroups.Groups.Remove(strconv.Itoa(group.ID))
}

func SwitchOwner(groupId, logged string, adr std.Address) {
	group := organizer.GetGroupFromId(groupId)

	if group.Owner.Owner() != std.Address(logged) {
		panic("only the owner give ownership away")
	}

	ok := group.DAO.IsAdmin(adr)
	if !ok {
		panic("the new owner must be an admin")
	}

	group.Owner = ownable.NewWithAddress(adr)
	organizer.ListGroups.Groups.Set(strconv.Itoa(group.ID), group)
}

// Wrapper to vote on proposal
// Verify if the vote ID exist and if the vote's choice exists
func VoteOnProposal(groupId, voteIdStr, voteStr, logged string) {
	group := organizer.GetGroupFromId(groupId)

	voteId, err := strconv.Atoi(voteIdStr)
	if err != nil {
		panic("invalid vote id")
	}

	vote, err := strconv.Atoi(voteStr)
	if err != nil {
		panic("invalid vote choice")
	}

	if vote != organizer.VoteNay && vote != organizer.VoteYay {
		panic("vote should be yay (1) or nay (0)")
	}

	group.DAO.VoteProposal(voteId, vote, std.Address(logged))
}

// Display in a table the users awaiting approval into the group
// Let admin directly accept or refuse members
func displayPendingApproval(group organizer.FriendGroup, awaiting rotree.IReadOnlyTree, logged string) string {
	table := mdtable.Table{
		Headers: []string{"Address", "Inspect"},
	}

	awaiting.Iterate("", "", func(indexStr string, val any) bool {
		adr := std.Address(indexStr)

		table.Append([]string{adr.String(),
			md.Link("Approve", txlink.Call("AcceptAwaitingMember", "groupId", strconv.Itoa(group.ID), "logged", logged, "adr", adr.String())) + " | " +
				md.Link("Reject", txlink.Call("RefuseAwaitingMember", "groupId", strconv.Itoa(group.ID), "logged", logged, "adr", adr.String())),
		})
		return false
	})

	return table.String()
}

// Display the group current members
// Let the admin promote / demote / remove member
func displayMembers(group organizer.FriendGroup, whitelist rotree.IReadOnlyTree, logged string) string {
	table := mdtable.Table{
		Headers: []string{"Address", "Rank", "Action"},
	}

	whitelist.Iterate("", "", func(indexStr string, val any) bool {
		var actionAdmin string

		adr := std.Address(indexStr)
		if adr == std.Address(logged) {
			return false
		}

		var rank string
		switch {
		case adr == group.Owner.Owner():
			rank = "Owner"
		case group.DAO.IsAdmin(adr):
			rank = "Admin"
		default:
			rank = "Member"
		}

		if rank != "Owner" {
			if group.DAO.IsAdmin(adr) {
				actionAdmin = md.Link("Remove Admin", txlink.Call("RemoveAdminRank", "groupId", strconv.Itoa(group.ID), "logged", logged, "adr", adr.String())) + " | "
			} else {
				actionAdmin = md.Link("Make Admin", txlink.Call("MakeAdmin", "groupId", strconv.Itoa(group.ID), "logged", logged, "adr", adr.String())) + " | "
			}
		}

		table.Append([]string{adr.String(), rank,
			actionAdmin +
				md.Link("Kick out", txlink.Call("RemoveMember", "groupId", strconv.Itoa(group.ID), "logged", logged, "adr", adr.String())),
		})

		return false
	})

	return table.String()
}

// Display admin board with Waitlist and Whiteliste
// Let admins create new proposals
func displayAdminBoard(group organizer.FriendGroup, logged string) string {
	var out string

	out += md.H2("Admin Board") + "\n"

	out += md.Link("Create Proposal", txlink.Call("CreateProposals",
		"groupId", strconv.Itoa(group.ID), "title", "Default Title",
		"desc", "Default Desc",
		"callback", "Default Callback",
		"callbackArgs", "Default Callback Arguments",
		"logged", logged)) + "\n"

	out += md.H3("Approval Waitlist") + "\n"

	awaiting := group.DAO.GetAwaiting()
	if awaiting.Size() > 0 {
		out += displayPendingApproval(group, awaiting, logged)
	} else {
		out += md.Paragraph("No member waiting for approval") + "\n"
	}
	out += md.HorizontalRule() + "\n"

	out += md.H3("Members") + "\n"

	whitelist := group.DAO.GetWhitelisted()
	if whitelist.Size() > 1 {
		out += displayMembers(group, whitelist, logged)
	} else {
		out += md.Paragraph("You're the only member for now.") + "\n"
	}
	out += md.HorizontalRule() + "\n"

	return out
}

func displayHangouts(group organizer.FriendGroup) string {
	hangouts := group.DAO.GetHangouts()

	table := mdtable.Table{
		Headers: []string{"ID", "Date", "Time", "Place", "Description"},
	}

	hangouts.Iterate("", "", func(_ string, val any) bool {
		hangout := val.(dao.HangoutInfo)
		table.Append([]string{
			strconv.Itoa(hangout.Id),
			hangout.Date,
			hangout.Time,
			hangout.Place,
			hangout.Description,
		})

		return false
	})

	return table.String()
}

func displayProposals(group organizer.FriendGroup, activeOnly bool, logged string) string {
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
					"vote", strconv.Itoa(organizer.VoteYay),
					"logged", logged)) + " | " +
					md.Link("Nay", txlink.Call("VoteOnProposal",
						"groupId", strconv.Itoa(group.ID),
						"index", index,
						"vote", strconv.Itoa(organizer.VoteNay),
						"logged", logged))

			if group.DAO.IsAdmin(std.Address(logged)) {
				closeLink := md.Link("Close", txlink.Call("AdminCloseProposal",
					"groupId", strconv.Itoa(group.ID), "voteId", index, "logged", logged))
				if voteActions == "-" {
					voteActions = closeLink
				} else {
					voteActions += " | " + closeLink
				}
			}
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
	logged := parseQueryAddress(url)

	out += md.H1("Group: " + group.Title)
	out += md.H2("Created by: " + organizer.ResolveAddressName(group.Owner.Owner()))
	out += md.Paragraph("Number of friends: "+strconv.Itoa(group.DAO.Whitelist.Size())) + "\n"
	out += md.HorizontalRule()

	if group.DAO.IsAdmin(std.Address(logged)) {
		out += displayAdminBoard(group, logged)
	}

	out += md.H2("Hangouts")
	out += displayHangouts(group)
	out += md.HorizontalRule()

	out += md.H2("Active DAOs")
	out += displayProposals(group, true, logged)
	out += md.HorizontalRule()

	out += md.H2("Closed DAOs")
	out += displayProposals(group, false, logged)
	out += md.HorizontalRule()

	out += md.Paragraph(md.Link("Go back", "/r/chaegnal/organizer"))

	if group.Owner.Owner() == std.Address(logged) {
		out += md.Paragraph(" | " + md.Link("Delete Group", txlink.Call("DeleteFriendGroup", "groupId", strconv.Itoa(group.ID), "logged", logged)))
		out += md.Paragraph(" | " + md.Link("Give ownership", txlink.Call("SwitchOwner", "groupId", strconv.Itoa(group.ID), "logged", logged, "adr", "default")))
	}
	return out
}
