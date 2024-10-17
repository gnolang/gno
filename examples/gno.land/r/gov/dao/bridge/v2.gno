package bridge

import (
	"gno.land/p/demo/dao"
	"gno.land/p/demo/membstore"
	govdao "gno.land/r/gov/dao/v2"
)

// govdaoV2 is a wrapper for interacting with the /r/gov/dao/v2 Realm
type govdaoV2 struct{}

func (g *govdaoV2) Propose(request dao.ProposalRequest) uint64 {
	return govdao.Propose(request)
}

func (g *govdaoV2) VoteOnProposal(id uint64, option dao.VoteOption) {
	govdao.VoteOnProposal(id, option)
}

func (g *govdaoV2) ExecuteProposal(id uint64) {
	govdao.ExecuteProposal(id)
}

func (g *govdaoV2) GetPropStore() dao.PropStore {
	return govdao.GetPropStore()
}

func (g *govdaoV2) GetMembStore() membstore.MemberStore {
	return govdao.GetMembStore()
}

func (g *govdaoV2) NewGovDAOExecutor(cb func() error) dao.Executor {
	return govdao.NewGovDAOExecutor(cb)
}

func (g *govdaoV2) NewMemberPropExecutor(cb func() []membstore.Member) dao.Executor {
	return govdao.NewMemberPropExecutor(cb)
}

func (g *govdaoV2) NewMembStoreImplExecutor(cb func() membstore.MemberStore) dao.Executor {
	return govdao.NewMembStoreImplExecutor(cb)
}
