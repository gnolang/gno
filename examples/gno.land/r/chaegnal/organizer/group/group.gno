package group

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl/rotree"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/commondao"
	"gno.land/p/sunspirit/md"

	"gno.land/r/chaegnal/organizer"
)

// displayPendingApproval displays in a table the users awaiting approval into the group
// It lets admin directly accept or refuse members
func displayPendingApproval(group organizer.FriendGroup, awaiting rotree.IReadOnlyTree, logged string) string {
	table := mdtable.Table{
		Headers: []string{"Address"},
	}

	awaiting.Iterate("", "", func(indexStr string, val any) bool {
		adr := std.Address(indexStr)

		table.Append([]string{adr.String()})
		return false
	})

	return table.String()
}

// displayMembers displays the group current members
// It lets the admin promote / demote / remove member
func displayMembers(group organizer.FriendGroup, logged string) string {
	table := mdtable.Table{
		Headers: []string{"Address", "Rank"},
	}

	members := group.DAO.Members()
	members.IterateByOffset(0, members.Size(), func(adr std.Address) bool {
		// var actionAdmin string

		if adr == std.Address(logged) {
			return false
		}

		var rank string
		switch {
		case adr == group.Owner.Owner():
			rank = "Owner"
		case isAdmin(group.ID.String(), std.Address(adr)):
			rank = "Admin"
		default:
			rank = "Member"
		}

		table.Append([]string{adr.String(), rank})

		return false
	})

	return table.String()
}

// displayAdminBoard displays admin board with Waitlist and Whiteliste
// It lets admins create new proposals
func displayAdminBoard(group organizer.FriendGroup, logged string) string {
	var out string

	out += md.H2("Admin Board") + "\n"

	out += md.Link("Create Proposal", txlink.Call("CreateProposals",
		"groupId", group.ID.String(), "title", "Default Title",
		"desc", "Default Desc",
		"callback", "Default Callback",
		"callbackArgs", "Default Callback Arguments",
		"logged", logged)) + "\n"

	out += md.H3("Available callbacks:") + "\n"
	out += "acceptAwaitingMember(your address - target address)" + "\n\n"
	out += "refuseAwaitingMember(your address - target address)" + "\n\n"
	out += "makeAdmin(your address - target address)" + "\n\n"
	out += "removeAdminRank(your address - target address)" + "\n\n"
	out += "removeMember(your address - target address)" + "\n\n"
	out += "SetHangoutTimeAndPlace(date - time - place - description)" + "\n\n"

	out += md.H3("Approval Waitlist") + "\n"

	if group.WaitList.Size() > 0 {
		out += displayPendingApproval(group, group.WaitList, logged)
	} else {
		out += md.Paragraph("No member waiting for approval") + "\n"
	}
	out += md.HorizontalRule() + "\n"

	out += md.H3("Members") + "\n"

	if group.DAO.Members().Size() > 1 {
		out += displayMembers(group, logged)
	} else {
		out += md.Paragraph("You're the only member for now.") + "\n"
	}
	out += md.HorizontalRule() + "\n"

	return out
}

func displayActiveProposals(group organizer.FriendGroup, logged string) string {
	if group.DAO.ActiveProposals().Size() <= 0 {
		return "No active proposals" + "\n\n"
	}

	table := mdtable.Table{
		Headers: []string{"ID", "Title", "Description", "Status", "Votes", "Actions", "Result"},
	}

	group.DAO.ActiveProposals().Iterate(0, group.DAO.ActiveProposals().Size(), false, func(p *commondao.Proposal) bool {
		record := p.VotingRecord()
		yay := record.VoteCount(commondao.ChoiceYes)
		nay := record.VoteCount(commondao.ChoiceNo)

		proposalID := strconv.FormatUint(p.ID(), 10)

		voteActions := md.Link("Yay", txlink.Call("VoteOnProposal",
			"groupId", group.ID.String(),
			"index", proposalID,
			"vote", strconv.Itoa(organizer.VoteYay),
			"logged", logged)) + " | " +
			md.Link("Nay", txlink.Call("VoteOnProposal",
				"groupId", group.ID.String(),
				"index", proposalID,
				"vote", strconv.Itoa(organizer.VoteNay),
				"logged", logged))

		if isAdmin(group.ID.String(), std.Address(logged)) {
			closeLink := md.Link("Check", txlink.Call("ExecuteExpiredProposals",
				"groupId", group.ID.String()))
			voteActions += " | " + closeLink
		}

		table.Append([]string{
			proposalID,
			p.Definition().Title(),
			p.Definition().Body(),
			"Active",
			"Yay: " + strconv.Itoa(yay) + " / Nay: " + strconv.Itoa(nay),
			voteActions,
			"-",
		})

		return false
	})

	return table.String()
}

func displayClosedProposals(group organizer.FriendGroup) string {
	if group.DAO.FinishedProposals().Size() <= 0 {
		return "No closed proposals yet" + "\n\n"
	}

	table := mdtable.Table{
		Headers: []string{"ID", "Title", "Description", "Status", "Votes", "Actions", "Result"},
	}

	group.DAO.FinishedProposals().Iterate(0, group.DAO.FinishedProposals().Size(), false, func(p *commondao.Proposal) bool {
		record := p.VotingRecord()
		yay := record.VoteCount(commondao.ChoiceYes)
		nay := record.VoteCount(commondao.ChoiceNo)

		proposalID := strconv.FormatUint(p.ID(), 10)

		table.Append([]string{
			proposalID,
			p.Definition().Title(),
			p.Definition().Body(),
			"Closed",
			"Yay: " + strconv.Itoa(yay) + " / Nay: " + strconv.Itoa(nay),
			"-",
			string(p.Status()),
		})

		return false
	})

	return table.String()
}

// displayHangouts displays the validated hangouts
func displayHangouts(group organizer.FriendGroup) string {
	if group.Hangout.Size() <= 0 {
		return "No Hangouts yet" + "\n\n"
	}

	table := mdtable.Table{
		Headers: []string{"ID", "Date", "Time", "Place", "Description"},
	}

	group.Hangout.Iterate("", "", func(indexStr string, val any) bool {
		hangout, ok := val.(organizer.HangoutInfo)
		if !ok {
			return false
		}

		table.Append([]string{
			indexStr,
			hangout.Date,
			hangout.Time,
			hangout.Place,
			hangout.Description,
		})

		return false
	})

	return table.String()
}

func Render(url string) string {
	var out string

	group := getGroupFromUrl(url)
	logged := parseQueryKeyword(url, "address")

	out += md.H1("Group: " + group.Title)
	out += md.H2("Created by: " + organizer.NameOrAddress(group.Owner.Owner()))
	out += md.Paragraph("Number of friends: "+strconv.Itoa(group.DAO.Members().Size())) + "\n"
	out += md.HorizontalRule()

	if isAdmin(group.ID.String(), std.Address(logged)) {
		out += displayAdminBoard(group, logged)
	}

	out += md.H2("Hangouts")
	out += displayHangouts(group)
	out += md.HorizontalRule()

	out += md.H2("Active Proposals")
	out += displayActiveProposals(group, logged)
	out += md.HorizontalRule()

	out += md.H2("Closed Proposals")
	out += displayClosedProposals(group)
	out += md.HorizontalRule()

	out += md.Paragraph(md.Link("Go back", "/r/chaegnal/organizer"))

	if group.Owner.Owner() == std.Address(logged) {
		out += md.Paragraph(" | " + md.Link("Delete Group", txlink.Call("DeleteFriendGroup", "groupId", group.ID.String(), "logged", logged)))
		out += md.Paragraph(" | " + md.Link("Give ownership", txlink.Call("SwitchOwner", "groupId", group.ID.String(), "logged", logged, "adr", "default")))
	}

	return out
}

// Helpers

// parseQueryKeyword parses the URL to retrieve the data behind the "=" sign according to the keyword given
func parseQueryKeyword(url, keyword string) string {
	query := strings.SplitAfter(url, "?")[1]
	params := strings.Split(query, "&")

	for _, param := range params {
		if strings.HasPrefix(param, keyword+"=") {
			return strings.SplitAfter(param, "=")[1]
		}
	}
	return ""
}

// getGroupFromUrl retrieves the group ID from the URL
func getGroupFromUrl(url string) organizer.FriendGroup {
	groupId := parseQueryKeyword(url, "groupId")
	if groupId == "" {
		panic("no group id provided")
	}

	return organizer.GetGroupFromId(groupId)
}
