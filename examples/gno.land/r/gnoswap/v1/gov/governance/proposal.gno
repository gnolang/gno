package governance

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"

	"gno.land/r/gnoswap/v1/common"
	en "gno.land/r/gnoswap/v1/emission"
	"gno.land/r/gnoswap/v1/halt"
)

// Governance can execute multiple messages in a single proposal
// To separate the messages, we use this separator
// Look for ProposeParameterChange() for more details
const (
	GOV_SPLIT   = "*GOV*"
	ROLE_PREFIX = "ROLE_CHANGE:"
)

// ProposeText creates a new text proposal with the given data
// It checks if the proposer is eligible to create a proposal and if they don't have an active proposal.
// Returns the proposal ID
// ref: https://docs.gnoswap.io/contracts/governance/proposal.gno#proposetext
func ProposeText(
	cur realm,
	title string,
	description string,
) (newProposalId uint64) {
	if err := halt.IsHalted(); err != nil {
		panic(err)
	}

	en.MintAndDistributeGns(cross)

	// Use state transition to create proposal
	proposer := std.PreviousRealm().Address()
	transition := CreateProposalTransition(proposer, Text, title, description)
	if err := applyTransition(transition); err != nil {
		panic(err)
	}

	// Get the new proposal ID from the updated state
	newProposalId = governanceState.GetProposalCounter()

	// Get the created proposal for event data
	proposal, exists := governanceState.GetProposal(newProposalId)
	if !exists {
		panic(ufmt.Sprintf("proposal(%d) not found", newProposalId))
	}

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"ProposeText",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"title", title,
		"description", description,
		"proposalId", formatUint(newProposalId),
		"quorumAmount", formatUint(proposal.QuorumAmount),
		"maxVotingWeight", proposal.MaxVotingWeight.ToString(),
		"configVersion", formatUint(proposal.ConfigVersion),
		"createdAt", formatUint(proposal.State.CreatedAt),
	)
	return newProposalId
}

// ProposeCommunityPoolSpend creates a new community pool spend proposal with the given data
// It checks if the proposer is eligible to create a proposal and if they don't have an active proposal.
// Returns the proposal ID
// ref: https://docs.gnoswap.io/contracts/governance/proposal.gno#proposecommunitypoolspend
func ProposeCommunityPoolSpend(
	cur realm,
	title string,
	description string,
	to std.Address,
	tokenPath string,
	amount uint64,
) (newProposalId uint64) {
	// Skip halt check if current level is MainnetSafeMode
	currentLevel := halt.GetCurrentHaltLevel()
	if currentLevel != halt.LvMainnetSafeMode {
		// Check if withdrawals are specifically halted, not just if the system is halted
		if err := halt.IsHalted(halt.OpTypeWithdraw); err != nil {
			panic(err.Error())
		}

		if halt.IsContractHalted(halt.OpTypeGovernance) {
			panic("gov/governance contract operations are currently disabled")
		}
	}

	// validate input
	if !to.IsValid() {
		panic(addDetailToError(
			errInvalidAddress,
			ufmt.Sprintf("address is %s", to.String()),
		))
	}
	common.MustRegistered(tokenPath)
	if amount == 0 {
		panic(addDetailToError(
			errInvalidInput,
			ufmt.Sprintf("amount is 0"),
		))
	}

	en.MintAndDistributeGns(cross)

	// Create base proposal
	proposer := std.PreviousRealm().Address()
	transition := CreateProposalTransition(proposer, CommunityPoolSpend, title, description)
	if err := applyTransition(transition); err != nil {
		panic(err)
	}

	// Get the new proposal ID from the updated state
	newProposalId = governanceState.GetProposalCounter()

	// Get the created proposal and update it with CommunityPoolSpend specific fields
	proposal, exists := governanceState.GetProposal(newProposalId)
	if !exists {
		panic(ufmt.Sprintf("proposal(%d) not found", newProposalId))
	}

	executeMessage := makeExecuteMessage(
		consts.COMMUNITY_POOL_PATH,
		"TransferToken",
		[]string{tokenPath, to.String(), formatUint(amount)},
	)

	proposal.CommunityPoolSpend = CommunityPoolSpendInfo{
		To:        to,
		TokenPath: tokenPath,
		Amount:    amount,
	}
	proposal.Execution = ExecutionInfo{
		Num:  1,
		Msgs: []string{executeMessage},
	}

	// Update the proposal in state
	governanceState = governanceState.WithUpdatedProposal(newProposalId, func(p ProposalInfo) ProposalInfo {
		return proposal
	})

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"ProposeCommunityPoolSpend",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"title", title,
		"description", description,
		"to", to.String(),
		"tokenPath", tokenPath,
		"amount", formatUint(amount),
		"proposalId", formatUint(newProposalId),
		"quorumAmount", formatUint(proposal.QuorumAmount),
		"maxVotingWeight", proposal.MaxVotingWeight.ToString(),
		"configVersion", formatUint(proposal.ConfigVersion),
		"createdAt", formatUint(proposal.State.CreatedAt),
	)

	return newProposalId
}

// ProposeParameterChange creates a new parameter change with the given data
// It checks if the proposer is eligible to create a proposal and if they don't have an active proposal.
// Through gnoswap interface, following string will be passed in for executions:
// <pkgPath>*EXE*<function>*EXE*<params>*GOV*<pkgPath>*EXE*<function>*EXE*<params>
// *GOV* is used to separate the messages
//
// Returns the proposal ID
// ref: https://docs.gnoswap.io/contracts/governance/proposal.gno#proposeparameterchange
func ProposeParameterChange(
	cur realm,
	title string,
	description string,
	numToExecute uint64,
	executions string,
) (newProposalId uint64) {
	if err := halt.IsHalted(); err != nil {
		panic(err)
	}

	// validate input
	if numToExecute == 0 {
		panic(addDetailToError(
			errInvalidInput,
			ufmt.Sprintf("numToExecute is 0"),
		))
	}

	proposer := std.PreviousRealm().Address()
	if strings.HasPrefix(executions, ROLE_PREFIX) {
		return handleRoleChangeProposal(cur, title, description, proposer, executions)
	}

	// check if numToExecute is a valid number
	splitGov := strings.Split(executions, GOV_SPLIT)
	if uint64(len(splitGov)) != numToExecute {
		panic(addDetailToError(
			errInvalidInput,
			ufmt.Sprintf("numToExecute(%d) does not match the number of executions(%d)", numToExecute, len(splitGov)),
		))
	}

	en.MintAndDistributeGns(cross)

	transition := CreateProposalTransition(proposer, ParameterChange, title, description)
	if err := applyTransition(transition); err != nil {
		panic(err)
	}

	newProposalId = governanceState.GetProposalCounter()
	proposal, exists := governanceState.GetProposal(newProposalId)
	if !exists {
		panic(ufmt.Sprintf("proposal(%d) not found", newProposalId))
	}

	// Set execution info for parameter change
	proposal.Execution = ExecutionInfo{
		Num:  numToExecute,
		Msgs: splitGov,
	}

	// Update the proposal in state with execution info
	governanceState = governanceState.WithUpdatedProposal(newProposalId, func(p ProposalInfo) ProposalInfo {
		return proposal
	})

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"ProposeParameterChange",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"title", title,
		"description", description,
		"numToExecute", formatUint(numToExecute),
		"executions", executions,
		"proposalId", formatUint(newProposalId),
		"quorumAmount", formatUint(proposal.QuorumAmount),
		"maxVotingWeight", proposal.MaxVotingWeight.ToString(),
		"configVersion", formatUint(proposal.ConfigVersion),
		"createdAt", formatUint(proposal.State.CreatedAt),
	)

	return newProposalId
}
