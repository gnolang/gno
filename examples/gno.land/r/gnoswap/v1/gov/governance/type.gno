package governance

import (
	"std"

	u256 "gno.land/p/gnoswap/uint256"
)

// Config represents the configuration of the governor contract
// ref: https://docs.gnoswap.io/core-concepts/governance#scope-of-proposals-5
type Config struct {
	// How long after a proposal is created does voting starrt
	VotingStartDelay uint64

	// The period during which votes are collected
	VotingPeriod uint64

	// Over how many seconds the voting weight is averaged for proposal voting as creation/cancellation threshold
	VotingWeightSmoothingDuration uint64

	// Percentags of x gns total supply at the time of proposal creation
	Quorum uint64

	// The munimum amount of average votes required to create a proposal
	ProposalCreationThreshold uint64

	// How much time must pass after the end of a voting period before a proposal can be executed
	ExecutionDelay uint64

	// The amount of time after the execution delay that the proposal can be executed
	ExecutionWindow uint64
}

// ProposalState represents the state of a proposal's execution
type ProposalState struct {
	Created   bool
	CreatedAt uint64

	Upcoming bool
	Active   bool

	Passed   bool
	PassedAt uint64

	Rejected   bool
	RejectedAt uint64

	Canceled   bool
	CanceledAt uint64

	// LABEL
	Executed   bool
	ExecutedAt uint64

	Expired   bool
	ExpiredAt uint64
}

type CommunityPoolSpendInfo struct {
	To        std.Address
	TokenPath string
	Amount    uint64
}

type ExecutionInfo struct {
	Num  uint64
	Msgs []string // split by *GOV*
}

type ParameterChangeInfo struct {
	PkgPath  string
	Function string
	Params   string
}

type ProposalType string

const (
	Text               ProposalType = "TEXT"
	CommunityPoolSpend ProposalType = "COMMUNITY_POOL_SPEND"
	ParameterChange    ProposalType = "PARAMETER_CHANGE"
	RoleChange         ProposalType = "ROLE_CHANGE"
)

func tryParseProposalType(v string) (ProposalType, error) {
	switch v {
	case "TEXT":
		return Text, nil
	case "COMMUNITY_POOL_SPEND":
		return CommunityPoolSpend, nil
	case "PARAMETER_CHANGE":
		return ParameterChange, nil
	case "ROLE_CHANGE":
		return RoleChange, nil
	default:
		return "", errInvalidProposalType
	}
}

func (p ProposalType) String() string {
	return string(p)
}

// ProposalInfo represents all the information about a proposal
type ProposalInfo struct {
	// The address of the proposer
	Proposer std.Address

	// Text, CommunityPoolSpend, ParameterChange
	ProposalType ProposalType

	// The execution state of the proposal
	State ProposalState

	// How many yes votes have been collected
	Yea *u256.Uint

	// How many no votes have been collected
	Nay *u256.Uint

	// The max voting weight at the time of proposal creation
	MaxVotingWeight *u256.Uint

	PossibleAddressWithWeight map[std.Address]uint64

	// The version of the config that this proposal was created with
	ConfigVersion uint64

	// How many total votes must be collected for the proposal
	QuorumAmount uint64

	// The title of the proposal
	Title string

	// The description of the proposal
	Description string

	// CommunityPoolSpend
	CommunityPoolSpend CommunityPoolSpendInfo

	// Execution (ParamterChange)
	Execution ExecutionInfo

	// RoleChange
	RoleChange RoleChangeInfo
}
