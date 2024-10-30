package bridge

import (
	"gno.land/p/demo/dao"
	"gno.land/p/demo/membstore"
)

// DAO abstracts the commonly used DAO interface
type DAO interface {
	Propose(dao.ProposalRequest) uint64
	VoteOnProposal(uint64, dao.VoteOption)
	ExecuteProposal(uint64)
	GetPropStore() dao.PropStore
	GetMembStore() membstore.MemberStore

	NewGovDAOExecutor(func() error) dao.Executor
}
