package evaldao

import (
	"gno.land/p/demo/gor/counters"
	governor "gno.land/p/demo/gor/governor"
	"gno.land/p/demo/maths"
	"gno.land/r/demo/groups"
	"std"
	"strconv"
)

// TODO: need a keeper for DAO, DAO=groups+governor
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

	// settings
	proposalThreshold := maths.NewRat(0, 1)
	quorum := maths.NewRat(1, 3)
	succeedThreshold := maths.NewRat(2, 3)
	vetoThreshold := maths.NewRat(1, 3)
	gs := governor.NewGovernorSettings(0, 5, proposalThreshold, quorum, succeedThreshold, vetoThreshold)

	gvr = governor.NewGovernor("eval_governor", votingPower, tally, gs)
}

// DAO miscs
func AddMember(addr std.Address, weight int, meta string) {
	groups.AddMember(groupID, addr.String(), weight, meta)
}

func RenderGroup() string {
	return groups.RenderGroup(groupID)
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

func Propose(propID string, title, detail, summary, url string, payee std.Address, amount std.Coins, propType string) string {
	std.AssertOriginCall()
	proposalID := gvr.Propose(propID, title, detail, summary, url, payee, amount, propType)
	return proposalID
}

func CastVote(propID string, vt string) error {
	std.AssertOriginCall()
	if err := gvr.CastVote(propID, vt); err != nil {
		return err
	}
	return nil
}

func Execute(propID string) {
	std.AssertOriginCall()
	gvr.Execute(propID)
}
