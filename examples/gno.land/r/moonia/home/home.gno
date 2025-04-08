package home

import (
	"gno.land/p/moul/txlink"
	"gno.land/p/moul/md"
)

func countActive() int {
	count := 0
	for _, p := range ds.Proposals.Proposals {
		if p.Active {
			count++
		}
	}
	return count
}

func renderDAO() string {
	out := "# Moonia's DAO\n\n"
	out += ds.DAO.ShowAdmin() + "\n\n"
	out += ds.DAO.Stats(len(ds.Proposals.Proposals), countActive()) + "\n"
	out += "## Actions:\n\n"
	if ds.DAO.Admin == "" {
		out += "- " + md.Link("Set Admin (once)", txlink.Call("SetAdmin")) + "\n"
	}
	out += "- " + md.Link("Join DAO", txlink.Call("JoinDAO")) + "\n"
	out += "- " + md.Link("Leave DAO", txlink.Call("LeaveDAO")) + "\n"
	out += "- " + md.Link("Create Sample Proposal", txlink.Call("CreateProposalTest")) + "\n"
	out += ds.DAO.ShowWhitelist() + "\n"
	out += ds.Proposals.ShowProposals()
	return out
}

func Render(path string) string {
	return renderDAO()
}
