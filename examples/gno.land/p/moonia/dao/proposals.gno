package dao

import (
	"std"
	"strconv"
	"gno.land/p/moonia/utils"
	"gno.land/p/moul/txlink"
	"gno.land/p/moul/md"
)

type Proposal struct {
	Title       string
	Description string
	Creator     std.Address
	YesVotes    int
	NoVotes     int
	Voters      map[std.Address]bool
	Active      bool
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

func (ps *ProposalStore) CreateProposal(title string, description string) string {
	caller := utils.GetCaller()
	if !ps.DAO.Whitelist[caller] {
		panic("Only whitelisted members can create proposals.")
	}
	p := Proposal{
		Title:       title,
		Description: description,
		Creator:     caller,
		Voters:      make(map[std.Address]bool),
		Active:      true,
	}
	ps.Proposals = append(ps.Proposals, p)
	return "Proposal created: " + title
}

func (ps *ProposalStore) CreateProposalTest() string {
	return ps.CreateProposal("Survey", "Would you like to visit Guatemala?")
}

func (ps *ProposalStore) CloseProposal(indexStr string) string {
	index := utils.ParseIndex(indexStr, len(ps.Proposals))
	p := &ps.Proposals[index]
	if p.Creator != utils.GetCaller() {
		panic("Only the proposal creator can close it.")
	}
	if !p.Active {
		panic("Proposal is already closed.")
	}
	p.Active = false
	return "Proposal '" + p.Title + "' has been closed."
}

// gnokey maketx call -pkgpath "gno.land/r/moonia/home" -func "Vote" -args "0" -args "true" -gas-fee 1000000ugnot -gas-wanted 5000000 -broadcast -chainid "dev" -remote "tcp://127.0.0.1:26657" g15dz69sch7fkhc9gk57hpe4qea77thmy20apu9x
func (ps *ProposalStore) Vote(indexStr string, voteYesStr string) string {
	index := utils.ParseIndex(indexStr, len(ps.Proposals))
	voteYes := voteYesStr == "true"
	caller := utils.GetCaller()

	if !ps.DAO.Whitelist[caller] {
		panic("Only whitelisted members can vote.")
	}

	p := &ps.Proposals[index]
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

// gnokey maketx call -pkgpath "gno.land/r/moonia/home" -func "EditProposal" -args "2" -args "hehehehe" -args "hihihihi" -gas-fee 1000000ugnot -gas-wanted 5000000 -broadcast -chainid "dev" -remote "tcp://127.0.0.1:26657" g15dz69sch7fkhc9gk57hpe4qea77thmy20apu9x
func (ps *ProposalStore) EditProposal(indexStr, newTitle, newDescription string) string {
	index := utils.ParseIndex(indexStr, len(ps.Proposals))
	p := &ps.Proposals[index]
	if p.Creator != utils.GetCaller() {
		panic("Only the creator can edit the proposal.")
	}
	if !p.Active {
		panic("Cannot edit a closed proposal.")
	}
	p.Title = newTitle
	p.Description = newDescription
	return "Proposal updated: " + newTitle + newDescription
}

// gnokey maketx call -pkgpath "gno.land/r/moonia/home" -func "GetProposal" -args "0" -gas-fee 1000000ugnot -gas-wanted 5000000 -broadcast -chainid "dev" -remote "tcp://127.0.0.1:26657" g15dz69sch7fkhc9gk57hpe4qea77thmy20apu9x
func (ps *ProposalStore) GetProposal(indexStr string) Proposal {
	index := utils.ParseIndex(indexStr, len(ps.Proposals))
	return ps.Proposals[index]
}
