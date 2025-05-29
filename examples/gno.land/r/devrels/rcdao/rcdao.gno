package rcdao

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
	"gno.land/p/nt/commondao"
	"gno.land/p/sunspirit/md"
)

var dao *commondao.CommonDAO

func init() {
	dao = commondao.New(
		commondao.WithName("Regional Community DAO"),
		commondao.WithMember("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"))
}

// I want to have a quick view of members
// I want to have an easy render, name it DefaultRender
// Exposing RO trees for members & props would make it easy to integrate with existing p/

func ProposeMeetup(title, body, location string, reqBudget std.Coins, beneficiary std.Address) uint64 {
	crossing()

	proposer := std.PreviousRealm().Address()
	if !dao.Members().Has(proposer) {
		panic("only members can propose")
	}

	prop := EventProposal{
		title:           title,
		body:            body,
		location:        location,
		requestedBudget: reqBudget,
		beneficiary:     beneficiary,
	}

	p := dao.MustPropose(proposer, prop)
	std.Emit("NewProp", "id", strconv.Itoa(int(p.ID())), "title", p.Definition().Title())

	return p.ID()
}

func Render(_ string) string {
	out := md.H1(dao.Name())

	out += "---\n\n"
	out += "## Active Props\n\n"
	out += ufmt.Sprintf("Active prop num: %d", dao.ActiveProposals().Size())

	dao.ActiveProposals().Iterate(0, 0, false, func(prop *commondao.Proposal) bool {
		out += prop.Definition().Title()
		return false
	})

	return out
}
