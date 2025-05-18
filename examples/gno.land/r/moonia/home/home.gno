package home

import (
	"std"
	"strconv"

	"gno.land/p/moul/txlink"
	"gno.land/p/moul/md"
	"gno.land/p/moonia/dao"
	"gno.land/p/moonia/utils"
)

var (
	ds DAOState
	daoMap map[string]*dao.DAO
	proposalMap map[string]*dao.ProposalStore
)

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

func CreateProposalSample(daoID string) string {
	d := daoMap[daoID]
	if d == nil {
		panic("DAO not found: " + daoID)
	}
	ds.DAO = d
	ds.Proposals = proposalMap[daoID]
	if ds.Proposals == nil {
		panic("ProposalStore not found for DAO: " + daoID)
	}
	return ds.Proposals.CreateProposal("Survey", "Would you like to visit Guatemala?", 30)
}

func renderDAO(daoID string) string {
	// TODO: fix empty string for caller, for the moment -> change it manually
	caller := std.Address("g15dz69sch7fkhc9gk57hpe4qea77thmy20apu9x")
	trueCaller := std.PreviousRealm().Address() // result: empty string
	out := "# " + ds.DAO.Name + "\n\n"
	out += "_" + ds.DAO.Description + "_\n\n"
	out += ds.DAO.ShowAdmin() + "\n\n"
	out += "Caller (debug): `" + trueCaller.String() + "`\n"

	out += "## Actions:\n\n"
	if !ds.DAO.IsMember(caller) && ds.DAO.Admin != "" {
		out += "- " + md.Link("Request to Join DAO", txlink.Call("RequestDAO")) + "\n"
	}	
	out += "- " + md.Link("Leave DAO", txlink.Call("LeaveDAO")) + "\n"
	out += "- " + md.Link("Create Sample Proposal", txlink.Call("CreateProposalSample", "args", "0")) + "\n"
	out += "- " + md.Link("View Stats", "/r/moonia/home?dao="+daoID+"&stats") + "\n"
	out += "- " + md.Link("Transfer Admin", txlink.Call("TransferAdmin", "args", "addr")) + "\n"
	out += ds.DAO.ShowWhitelist() + "\n"
	if ds.DAO.IsAdmin(caller) {
		out += "\n## 📨 Join Requests: " + strconv.Itoa(len(ds.DAO.Requests)) + "\n"
		if len(ds.DAO.Requests) == 0 {
			out += "_No pending requests._\n"
		} else {
			for addr := range ds.DAO.Requests {
				out += "- " + addr.String() + " " +
					md.Link("[Accept]", txlink.Call("AcceptRequest", "args", addr.String())) + " " +
					md.Link("[Decline]", txlink.Call("DeclineRequest", "args", addr.String())) + "\n"
			}
		}
	}	
	out += ds.Proposals.ShowProposals()
	return out
}

func Render(path string) string { 
	if path == "" {
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
	return renderDAO(daoID)
}
