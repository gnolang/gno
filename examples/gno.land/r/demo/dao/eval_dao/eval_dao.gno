package main

import (
	"gno.land/p/demo/gor/governor"
	"gno.land/p/demo/maths"
	"gno.land/r/demo/groups"
	"std"
)

// gvr is p package for build DAO, Gvr <-> DAO
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

	beforeExecute := func() error {
		println("before execute hook...")
		return nil
	}
	gvr.SetBeforeExecute(beforeExecute)

	afterExecute := governor.ExecuteHook(func() error {
		println("after execute hook...")
		println("here, after execution, tokens are transferred, need to update weight of members")
		// we need to get balance of member, and calculate a proper weight for them
		return nil
	})
	gvr.SetAfterExecute(afterExecute)
}

// Render
func RenderGvr(propID string) string {
	return gvr.RenderHome(propID)
}

// membership miscs
// TODO: calculate weight based on tokens (say)
func AddMember(addr std.Address, weight int, meta string) {
	std.AssertOriginCall()
	gvr.AddMember(addr.String(), weight, meta)
}

// // TODO: this should in gvr render, no need for here
func GetMemberVotingPower(mid uint64) int32 {
	std.AssertOriginCall()
	return gvr.GetVotingPowerFromMember(mid)
}

func GetTotalVotingPower() int32 {
	std.AssertOriginCall()
	return gvr.TotalVotingPower()
}

func UpdateMemberWeight(mid uint64, weight int) {
	gvr.UpdateMemberWeight(mid, weight)
}

// governor miscs
// propose -> vote -> tally -> execute -> render
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
