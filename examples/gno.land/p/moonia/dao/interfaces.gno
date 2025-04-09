package dao

import "std"

type IAdmin interface {
	SetAdmin() string
	TransferAdmin(addr string) string
	KickMember(addr std.Address) string
	AddMember(addr std.Address) string
}

type IDao interface {
	JoinDAO() string
	LeaveDAO() string
	ShowWhitelist() string
	Stats(totalProposals, activeProposals int) string
}

type IProposal interface {
	CreateProposal(title string, description string) string
	CloseProposal(indexStr string) string
	Vote(indexStr string, voteYesStr string) string
	EditProposal(indexStr, newTitle, newDescription string) string
	GetProposal(indexStr string) Proposal
	ShowProposals() string
	// SetVotingPeriod(duration string) string // by the proposal owner (maybe admin too)
}
