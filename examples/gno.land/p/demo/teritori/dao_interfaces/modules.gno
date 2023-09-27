package dao_interfaces

import (
	"std"
)

type ModuleInfo struct {
	Kind    string
	Version string
}

// NOTE: Some queries take a height param in DA0-DA0 contracts, but since gno seem to aim to support queries at any height, we shouldn't need it

type IVotingModule interface {
	Info() ModuleInfo
	ConfigJSON() string
	Render(path string) string
	VotingPowerAtHeight(address std.Address, height int64) (power uint64)
	TotalPowerAtHeight(height int64) uint64
}

type VotingModuleFactory func(core IDAOCore) IVotingModule

type IProposalModule interface {
	Core() IDAOCore
	Info() ModuleInfo
	ConfigJSON() string
	Render(path string) string
	Execute(proposalID int)
	VoteJSON(proposalID int, voteJSON string)
	ProposeJSON(proposalJSON string) int
	ProposalsJSON(limit int, startAfter string, reverse bool) string
	ProposalJSON(proposalID int) string
}

type ProposalModuleFactory func(core IDAOCore) IProposalModule
