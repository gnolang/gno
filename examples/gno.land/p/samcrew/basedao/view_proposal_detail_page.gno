package basedao

import (
	"strconv"

	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/seqid"
	"gno.land/p/nt/ufmt"
	"gno.land/p/samcrew/daocond"
	"gno.land/p/samcrew/daokit"
)

func (d *DAOPrivate) ProposalDetailPageView(idu uint64) string {
	return d.ProposalDetailHeaderView() +
		d.ProposalDetailView(idu)
}

func (d *DAOPrivate) ProposalDetailHeaderView() string {
	pkgPath := d.Realm.PkgPath()
	linkPath := getLinkPath(pkgPath)
	name := d.GetProfileString(d.Realm.Address(), "DisplayName", "DAO")
	s := ""
	s += ufmt.Sprintf("# %s - Proposal Detail\n\n", name)
	s += md.Link("> Go to Proposals", linkPath+":"+PROPOSALS_PATH) + "\n"
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

func (d *DAOPrivate) ProposalDetailView(idu uint64) string {
	id := seqid.ID(idu)
	proposal := d.Core.Proposals.GetProposal(uint64(id))
	s := ""
	s += ufmt.Sprintf("## Title - %s 📜\n\n", proposal.Title)
	s += ufmt.Sprintf("## Description 📝\n\n%s\n\n", proposal.Description)
	s += ufmt.Sprintf("## Resource - %s 📦\n\n", proposal.Action.Type())
	resource := d.Core.Resources.Get(proposal.Action.Type())
	s += ufmt.Sprintf("  - **Name:** %s\n", resource.DisplayName)
	s += ufmt.Sprintf("  - **Description:** %s\n", resource.Description)
	s += ufmt.Sprintf("  - **Condition:** %s\n\n", d.RenderWithRolesLinks(resource.Condition.Render()))
	s += ("---\n\n")
	s += proposal.Action.String() + "\n\n"
	s += ("---\n\n")
	proposal.UpdateStatus()
	if proposal.Status == daokit.ProposalStatusOpen {
		s += ufmt.Sprintf("## Status - Open 🟡\n\n")
		s += md.Link("Approve this proposal 🗳️", txlink.Call("Vote", "proposalID", strconv.FormatUint(idu, 10), "vote", daocond.VoteYes)) + "\n\n"
	} else if proposal.Status == daokit.ProposalStatusPassed {
		s += ufmt.Sprintf("## Status - Passed 🟢\n\n")
		s += md.Link("Execute this proposal 🗳️", txlink.Call("Execute", "proposalID", strconv.FormatUint(idu, 10))) + "\n\n"
	} else if proposal.Status == daokit.ProposalStatusExecuted {
		s += ufmt.Sprintf("## Status - Executed ✅\n\n")
	} else {
		s += ufmt.Sprintf("## Status - Closed 🔴\n\n")
	}

	signal := proposal.Condition.Signal(proposal.Ballot)
	s += ufmt.Sprintf("%f", signal) + "\n"
	s += ufmt.Sprintf("> proposed by %s 👤\n\n", proposal.ProposerID)
	s += ufmt.Sprintf("\n--------------------------------\n")
	s += ufmt.Sprintf("## Votes 🗳️\n\n%s\n\n", d.RenderWithRolesLinks(proposal.Condition.RenderWithVotes(proposal.Ballot)))
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}
