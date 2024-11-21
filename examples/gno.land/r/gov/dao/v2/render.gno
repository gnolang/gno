package govdao

import (
	"strconv"
	"strings"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/txlink"
	"gno.land/r/demo/users"
)

func Render(path string) string {
	var out string

	if path == "" {
		out += "# GovDAO Proposals\n\n"
		numProposals := d.Size()

		if numProposals == 0 {
			out += "No proposals found :(" // corner case
			return out
		}

		offset := uint64(0)
		if numProposals >= 10 {
			offset = uint64(numProposals) - 10
		}

		// Fetch the last 10 proposals
		proposals := d.Proposals(offset, uint64(10))
		for i := len(proposals) - 1; i >= 0; i-- {
			prop := proposals[i]

			title := prop.Title()
			if len(title) > 40 {
				title = title[:40] + "..."
			}

			propID := offset + uint64(i)
			out += ufmt.Sprintf("## [Prop #%d - %s](/r/gov/dao/v2:%d)\n\n", propID, title, propID)
			out += ufmt.Sprintf("**Status: %s**\n\n", strings.ToUpper(prop.Status().String()))

			user := users.GetUserByAddress(prop.Author())
			authorDisplayText := prop.Author().String()
			if user != nil {
				authorDisplayText = ufmt.Sprintf("[%s](/r/demo/users:%s)", user.Name, user.Name)
			}

			out += ufmt.Sprintf("**Author: %s**\n\n", authorDisplayText)

			if i != 0 {
				out += "---\n\n"
			}
		}

		return out
	}

	// Display the detailed proposal
	idx, err := strconv.Atoi(path)
	if err != nil {
		return "404: Invalid proposal ID"
	}

	// Fetch the proposal
	prop, err := d.ProposalByID(uint64(idx))
	if err != nil {
		return ufmt.Sprintf("unable to fetch proposal, %s", err.Error())
	}

	// Render the proposal page
	out += renderPropPage(prop, idx)

	return out
}

func renderPropPage(prop dao.Proposal, idx int) string {
	var out string

	out += ufmt.Sprintf("# Proposal #%d - %s\n\n", idx, prop.Title())
	out += prop.Render()
	out += renderAuthor(prop)
	out += renderActionBar(prop, idx)
	out += "\n\n"

	return out
}

func renderAuthor(p dao.Proposal) string {
	var out string

	authorUsername := ""
	user := users.GetUserByAddress(p.Author())
	if user != nil {
		authorUsername = user.Name
	}

	if authorUsername != "" {
		out += ufmt.Sprintf("**Author: [%s](/r/demo/users:%s)**\n\n", authorUsername, authorUsername)
	} else {
		out += ufmt.Sprintf("**Author: %s**\n\n", p.Author().String())
	}

	return out
}

func renderActionBar(p dao.Proposal, idx int) string {
	var out string

	out += "## Actions\n\n"
	if p.Status() == dao.Active {
		out += ufmt.Sprintf("[Vote YES](%s) - [Vote NO](%s) - [Vote ABSTAIN](%s)",
			txlink.URL("VoteOnProposal", "id", strconv.Itoa(idx), "option", "YES"),
			txlink.URL("VoteOnProposal", "id", strconv.Itoa(idx), "option", "NO"),
			txlink.URL("VoteOnProposal", "id", strconv.Itoa(idx), "option", "ABSTAIN"),
		)
	} else {
		out += "The voting period for this proposal is over.\n\n"
	}

	return out
}
