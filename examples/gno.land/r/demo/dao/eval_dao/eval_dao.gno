package eval_dao

import (
	"gno.land/p/demo/gor/governor"
	"gno.land/p/demo/maths"
	"gno.land/r/demo/dao/contrib_dao"
	"gno.land/r/demo/dao/gov_token"
	"gno.land/r/demo/groups"
	"std"
)

const zeroAddress = std.Address("")

// gvr is p package for build DAO, Gvr <-> DAO
var (
	gvr *governor.Governor
)

func CreateDAO(name string) {
	std.AssertOriginCall()
	_, groupID := groups.NewGroup(name)
	println("group created: ", groupID)

	membership := governor.NewGroupAdapter(groupID)

	// settings
	proposalThreshold := maths.NewRat(0, 1)
	quorum := maths.NewRat(1, 3)
	succeedThreshold := maths.NewRat(2, 3)
	vetoThreshold := maths.NewRat(1, 3)
	gs := governor.NewGovernorSettings(0, 5, proposalThreshold, quorum, succeedThreshold, vetoThreshold)

	gvr = governor.NewGovernor(name+"_governor", membership, gs)
	// until now, we have a gvr, consists of membership, and propose, vote abilities

	//beforeExecute := func() error {
	//	println("before execute hook...")
	//	return nil
	//}
	//gvr.SetBeforeExecute(beforeExecute)

	println("total supply of eval_dao: ", gov_token.GetTotalSupply())
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

func GetAmount(acc std.Address) uint64 {
	return gov_token.GetAmount(acc)
}

// governor miscs
// propose -> vote -> tally -> execute -> render
func Propose(title, detail, summary, url string, payee std.Address, amount std.Coins, propType string) string {
	std.AssertOriginCall()
	proposalID := gvr.Propose(title, detail, summary, url, payee, amount, propType)
	// set hook
	afterExecute := governor.ExecuteHook(func() error {
		println("after execution hook, tokens are transferred, need to update weight of members")
		// Transfer voting power
		gov_token.TransferVotingPower(zeroAddress, payee, 200000000) // TODO: for all coins, get amount
		// update weight rather than lazy read
		if gov_token.GetAmount(payee) > 1000 {
			// update weight
			println("update weight of 0 to 2")
			contrib_dao.UpdateMemberWeight(0, 2)
		}
		return nil
	})
	gvr.SetAfterExecute(afterExecute)

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
