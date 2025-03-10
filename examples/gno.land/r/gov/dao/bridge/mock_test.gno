package bridge

import (
	"gno.land/p/demo/dao"
	"gno.land/p/demo/membstore"
)

type (
	proposeDelegate           func(dao.ProposalRequest) uint64
	voteOnProposalDelegate    func(uint64, dao.VoteOption)
	executeProposalDelegate   func(uint64)
	getPropStoreDelegate      func() dao.PropStore
	getMembStoreDelegate      func() membstore.MemberStore
	newGovDAOExecutorDelegate func(func() error) dao.Executor
)

type mockDAO struct {
	proposeFn           proposeDelegate
	voteOnProposalFn    voteOnProposalDelegate
	executeProposalFn   executeProposalDelegate
	getPropStoreFn      getPropStoreDelegate
	getMembStoreFn      getMembStoreDelegate
	newGovDAOExecutorFn newGovDAOExecutorDelegate
}

func (m *mockDAO) Propose(request dao.ProposalRequest) uint64 {
	if m.proposeFn != nil {
		return m.proposeFn(request)
	}

	return 0
}

func (m *mockDAO) VoteOnProposal(id uint64, option dao.VoteOption) {
	if m.voteOnProposalFn != nil {
		m.voteOnProposalFn(id, option)
	}
}

func (m *mockDAO) ExecuteProposal(id uint64) {
	if m.executeProposalFn != nil {
		m.executeProposalFn(id)
	}
}

func (m *mockDAO) GetPropStore() dao.PropStore {
	if m.getPropStoreFn != nil {
		return m.getPropStoreFn()
	}

	return nil
}

func (m *mockDAO) GetMembStore() membstore.MemberStore {
	if m.getMembStoreFn != nil {
		return m.getMembStoreFn()
	}

	return nil
}

func (m *mockDAO) NewGovDAOExecutor(cb func() error) dao.Executor {
	if m.newGovDAOExecutorFn != nil {
		return m.newGovDAOExecutorFn(cb)
	}

	return nil
}
