package governance

func GetProposerByProposalId(proposalId uint64) string {
	return mustGetProposal(proposalId).Proposer.String()
}

func GetProposalTypeByProposalId(proposalId uint64) string {
	return mustGetProposal(proposalId).ProposalType.String()
}

func GetYeaByProposalId(proposalId uint64) string {
	return mustGetProposal(proposalId).Yea.ToString()
}

func GetNayByProposalId(proposalId uint64) string {
	return mustGetProposal(proposalId).Nay.ToString()
}

func GetConfigVersionByProposalId(proposalId uint64) uint64 {
	return mustGetProposal(proposalId).ConfigVersion
}

func GetQuorumAmountByProposalId(proposalId uint64) uint64 {
	return mustGetProposal(proposalId).QuorumAmount
}

func GetTitleByProposalId(proposalId uint64) string {
	return mustGetProposal(proposalId).Title
}

func GetDescriptionByProposalId(proposalId uint64) string {
	return mustGetProposal(proposalId).Description
}

func GetExecutionStateByProposalId(proposalId uint64) ProposalState {
	return mustGetProposal(proposalId).State
}
