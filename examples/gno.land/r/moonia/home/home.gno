package home

import (
	"gno.land/p/moul/txlink"
	"gno.land/p/moul/md"
	"gno.land/p/moonia/dao"
	"gno.land/p/moonia/utils"
)

var ds DAOState
var daoMap map[string]*dao.DAO
var proposalMap map[string]*dao.ProposalStore

type DAOState struct {
	DAO       *dao.DAO
	Proposals *dao.ProposalStore
}

func init() {
	daoMap = make(map[string]*dao.DAO)
	proposalMap = make(map[string]*dao.ProposalStore)
}

func countActive() int {
	count := 0
	for _, p := range ds.Proposals.Proposals {
		if p.Active {
			count++
		}
	}
	return count
}

func renderDAOStats() string {
	out := "# Statistics for `" + ds.DAO.Name + "`\n"
	totalProposals := len(ds.Proposals.Proposals)
	activeProposals := countActive()
	out += ds.DAO.Stats(totalProposals, activeProposals) + "\n"

	return out
}

func renderDAO() string {
	caller := utils.GetCaller()
	out := "# " + ds.DAO.Name + "\n\n"
	out += "_" + ds.DAO.Description + "_\n\n"
	out += ds.DAO.ShowAdmin() + "\n\n"

	out += "## Actions:\n\n"
	if !ds.DAO.IsMember(caller) && ds.DAO.Admin != "" {
		out += "- " + md.Link("Request to Join DAO", txlink.Call("RequestDAO")) + "\n"
	}
	out += "- " + md.Link("Leave DAO", txlink.Call("LeaveDAO")) + "\n"
	out += "- " + md.Link("Create Sample Proposal", txlink.Call("CreateProposalTest")) + "\n"
	for id := range daoMap {
		out += "- " + md.Link("View Stats", "/r/moonia/home?dao="+id+"&stats") + "\n"
	}
	out += "- " + md.Link("Transfer Admin", txlink.Call("TransferAdmin", "args", "addr")) + "\n"
	out += ds.DAO.ShowWhitelist() + "\n"
	out += ds.Proposals.ShowProposals()
	return out
}

func Render(path string) string {
	if path == "/" || path == "" {
		return renderDashboard()
	}
	daoID, showStats := utils.ParseQuery(path)
	d := daoMap[daoID]
	if d == nil {
		return "DAO not found: `" + daoID + "`"
	}
	ds.DAO = d
	ds.Proposals = proposalMap[daoID]

	if showStats {
		return renderDAOStats()
	}
	return renderDAO()
}
