package evaldao

import (
	"gno.land/p/demo/gor/counters"
	governor "gno.land/p/demo/gor/governor"
	"gno.land/r/demo/groups"
	"std"
	"strconv"
)

var (
	evalDAO *groups.Group
	groupID groups.GroupID
	gvr     *governor.Governor
	counter *counters.Counter
)

func init() {
	counter = counters.NewCounter()
}

// generate proposalId
func ProposalID() string {
	propID := counter.Current()
	counter.Increment()
	return strconv.FormatUint(propID, 10)
}

func CreateDAO() {
	std.AssertOriginCall()
	evalDAO, groupID = groups.NewGroup("eval_dao")
	println("group created: ", groupID)

	votingPower := governor.NewNoCoinAdapter(evalDAO)
	tally := governor.NewProposalVotes()
	gs := governor.NewGovernorSettings(0, 5, 0, 40, 67, 34)
	gvr = governor.NewGovernor("eval_governor", votingPower, tally, gs)

	// TODO:
	// render governor
	// render proposal
}

// DAO miscs
func AddMember(addr std.Address) {
	groups.AddMember(2, "g1vahx7atnv4erxh6lta047h6lta047h6ll85gpy", 55, "metadata3")
}

func TotalVotingPower() uint64 {
	//evalDAO.getNumMembers()
	return 0
}

func GetVotingPowerByAddr(member std.Address) uint64 {
	//return uint64(evalDAO.getMemberWeight(member))
	return 0
}

// TODO:
// render EvalDAO, groupID, num members, member, voting power
// remove member

// governor miscs
// propose -> vote -> tally -> render
func RenderGvr(propID string) string {
	return gvr.RenderHome(propID)
}

func Propose(propID string, title, detail, summary, url string, payee std.Address, propType string) string {
	std.AssertOriginCall()
	proposalID := gvr.Propose(propID, title, detail, summary, url, payee, propType)
	return proposalID
}

func CastVote(propID string, vt string) error {
	std.AssertOriginCall()
	if err := gvr.CastVote(propID, vt); err != nil {
		return err
	}
	return nil
}
