package dao

import (
	"std"
	"strconv"

	"gno.land/p/moonia/utils"
	"gno.land/p/moul/txlink"
	"gno.land/p/moul/md"
)

type Proposal struct {
	Title        string
	Description  string
	Creator      std.Address
	YesVotes     int
	NoVotes      int
	Voters       map[std.Address]bool
	Active       bool
	VotingPeriod int64
	CreatedAt    int64
}

type ProposalStore struct {
	Proposals []Proposal
	DAO       *DAO

}

func NewProposalStore(dao *DAO) *ProposalStore {
	return &ProposalStore{
		Proposals: []Proposal{},
		DAO:       dao,
	}
}

func (ps *ProposalStore) CreateProposal(title, description string, period int64) string {
	caller := std.PreviousRealm().Address()
	if !ps.DAO.Whitelist[caller] {
		panic("Only whitelisted members can create proposals.")
	}
	if period <= 0 {
		panic("Voting period must be greater than 0.")
	}

	p := Proposal{
		Title:        title,
		Description:  description,
		Creator:      caller,
		Voters:       make(map[std.Address]bool),
		Active:       true,
		VotingPeriod: period,
		CreatedAt:    utils.GetBlockTime(),
	}
	ps.Proposals = append(ps.Proposals, p)
	return "Proposal created: " + title + " Voting period: " + strconv.Itoa(int(period))
}

func (ps *ProposalStore) CloseProposal(indexStr string) string {
	index := utils.ParseIndex(indexStr, len(ps.Proposals))
	p := &ps.Proposals[index]
	if p.Creator != std.PreviousRealm().Address() {
		panic("Only the proposal creator can close it.")
	}
	if !p.Active {
		panic("Proposal is already closed.")
	}
	p.Active = false
	return "Proposal '" + p.Title + "' has been closed."
}

func (ps *ProposalStore) Vote(indexStr, voteYesStr string) string {
	index := utils.ParseIndex(indexStr, len(ps.Proposals))
	voteYes := voteYesStr == "true"
	caller := std.PreviousRealm().Address()

	if !ps.DAO.Whitelist[caller] {
		panic("Only whitelisted members can vote.")
	}

	p := &ps.Proposals[index]

	currentTime := utils.GetBlockTime()
	if currentTime - p.CreatedAt >= p.VotingPeriod {
		p.Active = false
		panic("Voting period is over. Proposal is now closed.")
	}

	if !p.Active {
		panic("Voting is closed for this proposal.")
	}
	if p.Voters[caller] {
		panic("You have already voted.")
	}

	p.Voters[caller] = true
	if voteYes {
		p.YesVotes++
		return "Vote recorded: YES for '" + p.Title + "'"
	} else {
		p.NoVotes++
		return "Vote recorded: NO for '" + p.Title + "'"
	}
}

func (ps *ProposalStore) ShowProposals() string {
	activeOut := "## Active Proposals\n\n"
	closedOut := "## Closed Proposals\n\n"
	hasActive := false
	hasClosed := false

	for i, p := range ps.Proposals {
		proposalStr := "**[" + strconv.Itoa(i) + "]** " + p.Title + "\n"
		proposalStr += p.Description + "\n\n"
		proposalStr += "by _" + p.Creator.String() + "_\n\n"
		proposalStr += "📅 Voting ends at: " + utils.FormatTimestamp(p.CreatedAt + p.VotingPeriod) + "\n\n"
		proposalStr += "✅ " + strconv.Itoa(p.YesVotes) + " | ❌ " + strconv.Itoa(p.NoVotes) + "\n"

		if p.Active {
			hasActive = true
			proposalStr += "(Active) — " +
				md.Link("Vote Yes", txlink.Call("Vote", "args", strconv.Itoa(i), "args", "true")) + " | " +
				md.Link("Vote No", txlink.Call("Vote", "args", strconv.Itoa(i), "args", "false")) + "\n\n" +
				md.Link("❌ Close proposal", txlink.Call("CloseProposal", "args", strconv.Itoa(i))) + "\n\n ---"
			activeOut += proposalStr + "\n\n"
		} else {
			hasClosed = true
			proposalStr += "(Closed)\n"
			closedOut += proposalStr + "\n\n ---- \n\n"
		}
	}
	if !hasActive {
		activeOut += "_No active proposals._\n\n"
	}
	if !hasClosed {
		closedOut += "_No closed proposals._\n\n"
	}
	return activeOut + "\n" + closedOut
}

func (ps *ProposalStore) EditProposal(indexStr, newTitle, newDescription string, newPeriod int64) string {
	index := utils.ParseIndex(indexStr, len(ps.Proposals))
	p := &ps.Proposals[index]
	if p.Creator != std.PreviousRealm().Address() {
		panic("Only the creator can edit the proposal.")
	}
	if !p.Active {
		panic("Cannot edit a closed proposal.")
	}
	if newPeriod <= 0 {
		panic("Voting period must be greater than 0.")
	}
	p.Title = newTitle
	p.Description = newDescription
	p.VotingPeriod = newPeriod
	return "Proposal updated: " + newTitle + newDescription + strconv.Itoa(int(newPeriod))
}

func (ps *ProposalStore) GetProposal(indexStr string) Proposal {
	index := utils.ParseIndex(indexStr, len(ps.Proposals))
	return ps.Proposals[index]
}