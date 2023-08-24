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

	// TODO: construct from parameters
	// settings
	proposalThreshold := maths.NewRat(0, 1)
	quorum := maths.NewRat(1, 3)
	succeedThreshold := maths.NewRat(2, 3)
	vetoThreshold := maths.NewRat(1, 3)
	gs := governor.NewGovernorSettings(0, 5, proposalThreshold, quorum, succeedThreshold, vetoThreshold)

	gvr = governor.NewGovernor(name+"_governor", membership, gs)
}

// Render
func Render(propID string) string {
	return gvr.RenderHome(propID)
}

// membership miscs
func AddMember(addr std.Address, weight int, meta string) {
	std.AssertOriginCall()
	gvr.AddMember(addr.String(), weight, meta)
}

// weight
func GetMemberVotingPower(addr std.Address) int32 {
	std.AssertOriginCall()
	return gvr.VotingPowerFromMember(addr)
}

// total weight
func GetTotalVotingPower() int32 {
	std.AssertOriginCall()
	return gvr.TotalVotingPower()
}

// tokens
func GetAmount(acc std.Address) uint64 {
	std.AssertOriginCall()
	return gov_token.GetAmount(acc)
}

// governor miscs
func Propose(title, detail, summary, url string, payee std.Address, amount std.Coins, propType string) string {
	std.AssertOriginCall()
	proposalID := gvr.Propose(title, detail, summary, url, payee, amount, propType)
	// set hook
	afterExecute := governor.ExecuteHook(func() error {
		println("after execution hook, tokens are transferred, need to update weight of members")
		// Transfer voting power
		amt := []std.Coin(amount)[0].Amount
		gov_token.TransferVotingPower(zeroAddress, payee, uint64(amt))
		// get amount of payee, and use this to update weights
		amount := gov_token.GetAmount(payee)
		contrib_dao.UpdateWeightCheckPoints(payee, int32(amount), 7) // TODO: consistency of all int32, uint64 for all tokens
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
