package basedao

import (
	"gno.land/p/moul/md"
	"gno.land/p/nt/ufmt"
	"gno.land/p/samcrew/daokit"
)

func (d *DAOPrivate) ProposalsPageView(path string) string {
	return d.ProposalsHeaderView() +
		d.ProposalsView(path)
}

func (d *DAOPrivate) ProposalsHeaderView() string {
	pkgPath := d.Realm.PkgPath()
	linkPath := getLinkPath(pkgPath)
	name := d.GetProfileString(d.Realm.Address(), "DisplayName", "DAO")
	s := ""
	s += ufmt.Sprintf("# %s - Proposals\n\n", name)
	s += md.Link("> Go to the Proposal history 📜", ufmt.Sprintf("%s:%s", linkPath, PROPOSAL_HISTORY_PATH)) + "\n\n"
	s += md.Link("Add a new proposal 🗳️", ufmt.Sprintf("%s$help", linkPath)) + "\n\n"
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

func (d *DAOPrivate) ProposalsView(path string) string {
	activeProposals := d.Core.Proposals.GetProposals(
		func(p daokit.Proposal) bool {
			return p.Status == daokit.ProposalStatusOpen || p.Status == daokit.ProposalStatusPassed
		},
	)
	s := ""
	s += ufmt.Sprintf("## Active Proposals 🗳️ (%d)\n\n", activeProposals.Tree.Size())
	s += d.RenderProposalsTable(path, activeProposals)
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}
