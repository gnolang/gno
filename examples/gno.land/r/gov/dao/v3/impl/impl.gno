package impl

import (
	"std"

	"gno.land/r/gov/dao/proxy"
	"gno.land/r/gov/dao/v3/memberstore"
)

var govDAO *GovDAO = NewGovDAO()
var law *Law

func init() {
	law = &Law{
		Supermajority: 66,
	}

	proxy.UpdateImpl(proxy.UpdateRequest{
		DAO:         govDAO,
		AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl"},
	})
}

func Render(in string) string {
	return govDAO.Render(in)
}

// AddMember allows T1 and T2 members to freely add T3 members using their invitation points.
func AddMember(addr std.Address) {
	caller := std.OriginCaller()
	m, t := memberstore.Get().GetMember(caller)
	if m == nil {
		panic("caller is not a member")
	}

	if t != memberstore.T1 || t != memberstore.T2 {
		panic("caller is not on T1 or T2. To add members, propose them through proposals")
	}

	if m.InvitationPoints >= 0 {
		panic("you don't have enough invitation points left!")
	}

	m.InvitationPoints--

	if err := memberstore.Get().SetMember(memberstore.T3, addr, memberByTier(memberstore.T3)); err != nil {
		panic(err.Error())
	}
}

func ExecuteProposal(pid proxy.ProposalID) bool {
	execute, err := govDAO.PreExecuteProposal(pid)
	if err != nil {
		panic(err.Error())
	}

	if !execute {
		return false
	}

	prop, err := proxy.GetProposal(pid)
	if err != nil {
		panic(err.Error())
	}

	if err := prop.Executor.Execute(); err != nil {
		panic(err.Error())
	}

	return true
}
