package basedao

import (
	"gno.land/p/moul/md"
	"gno.land/p/nt/ufmt"
	"gno.land/p/samcrew/daokit"
)

func (d *DAOPrivate) ProposalHistoryPageView(path string) string {
	return d.ProposalHistoryHeaderView() +
		d.ProposalHistoryView(path)
}

func (d *DAOPrivate) ProposalHistoryHeaderView() string {
	pkgPath := d.Realm.PkgPath()
	linkPath := getLinkPath(pkgPath)
	name := d.GetProfileString(d.Realm.Address(), "DisplayName", "DAO")
	s := ""
	s += ufmt.Sprintf("# %s - Proposal History\n\n", name)
	s += md.Link("> Go to the active Proposals 📃", ufmt.Sprintf("%s:%s", linkPath, PROPOSALS_PATH)) + "\n\n"
	s += md.Link("Add a new proposal 🗳️", ufmt.Sprintf("%s$help", linkPath)) + "\n\n"
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

func (d *DAOPrivate) ProposalHistoryView(path string) string {
	inactiveProposals := d.Core.Proposals.GetProposals(
		func(p daokit.Proposal) bool {
			return p.Status != daokit.ProposalStatusOpen && p.Status != daokit.ProposalStatusPassed
		},
	)
	s := ""
	s += ufmt.Sprintf("## Inactive Proposals 🗳️ (%d)\n\n", inactiveProposals.Tree.Size())
	s += d.RenderProposalsTable(path, inactiveProposals)
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}
