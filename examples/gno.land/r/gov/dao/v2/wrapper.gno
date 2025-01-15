package govdao

import (
	"gno.land/p/demo/dao"
	"gno.land/p/demo/membstore"
)

// daoWrapper wraps GovDAO to allow registering in the bridge
type daoWrapper struct{}

func (g *daoWrapper) Propose(request dao.ProposalRequest) uint64 {
	return Propose(request)
}

func (g *daoWrapper) VoteOnProposal(id uint64, option dao.VoteOption) {
	VoteOnProposal(id, option)
}

func (g *daoWrapper) ExecuteProposal(id uint64) {
	ExecuteProposal(id)
}

func (g *daoWrapper) GetPropStore() dao.PropStore {
	return GetPropStore()
}

func (g *daoWrapper) GetMembStore() membstore.MemberStore {
	return GetMembStore()
}

func (g *daoWrapper) NewGovDAOExecutor(cb func() error) dao.Executor {
	return NewGovDAOExecutor(cb)
}

func (g *daoWrapper) NewMemberPropExecutor(cb func() []membstore.Member) dao.Executor {
	return NewMemberPropExecutor(cb)
}

func (g *daoWrapper) NewMembStoreImplExecutor(cb func() membstore.MemberStore) dao.Executor {
	return NewMembStoreImplExecutor(cb)
}
