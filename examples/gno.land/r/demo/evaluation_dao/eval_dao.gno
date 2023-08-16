package evaldao

import (
	"gno.land/p/demo/gor/governor"
	"gno.land/p/demo/maths"
	"gno.land/r/demo/groups"
	"std"
)

// TODO: need a keeper for DAO, DAO=groups+governor
var (
	gvr *governor.Governor
)

func CreateDAO() {
	std.AssertOriginCall()
	_, groupID := groups.NewGroup("eval_dao")
	println("group created: ", groupID)

	membership := governor.NewNoCoinAdapter(groupID)

	// settings
	proposalThreshold := maths.NewRat(0, 1)
	quorum := maths.NewRat(1, 3)
	succeedThreshold := maths.NewRat(2, 3)
	vetoThreshold := maths.NewRat(1, 3)
	gs := governor.NewGovernorSettings(0, 5, proposalThreshold, quorum, succeedThreshold, vetoThreshold)

	gvr = governor.NewGovernor("eval_governor", membership, gs)
	// until now, we have a gvr, consists of membership, and propose, vote abilities
	// gvr is p package for build DAO, Gvr <-> DAO
}

// membership miscs
// should all assertOrigin
func AddMember(addr std.Address, weight int, meta string) {
	// TODO: calculate weight based on tokens (say)
	// need an update mechanism also
	gvr.Ms.AddMember(addr.String(), weight, meta)
}

// TODO: this should in gvr render, no need for here
func GetMemberVotingPower(mid uint64) int32 {
	return gvr.Ms.GetVotingPowerFromMember(mid)
}

func UpdateMemberWeight(mid uint64, weight int) {
	gvr.Ms.UpdateMemberWeight(mid, weight)
}

func GetTotalVotingPower() int32 {
	return gvr.Ms.TotalVotingPower()
}

// TODO:
// render EvalDAO, groupID, num members, member, voting power
// remove member

// governor miscs
// propose -> vote -> tally -> render
func RenderGvr(propID string) string {
	return gvr.RenderHome(propID)
}

func Propose(title, detail, summary, url string, payee std.Address, amount std.Coins, propType string) string {
	std.AssertOriginCall()
	proposalID := gvr.Propose(title, detail, summary, url, payee, amount, propType)
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
