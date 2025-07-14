package halt

var _ Operation = (*basicOperation)(nil)

const (
	rawPool          = "pool"
	rawPosition      = "position"
	rawProtocolFee   = "protocol_fee"
	rawRouter        = "router" // handle swap related operations
	rawStaker        = "staker" // handle liquidity related operations
	rawLaunchpad     = "launchpad"
	rawGovernance    = "governance" // gov/governance
	rawGovStaker     = "gov_staker" // gov/staker
	rawXGns          = "xgns"       // gov/xgns
	rawCommunityPool = "community_pool"
	rawEmission      = "emission"
	rawWithdraw      = "withdraw"
)

// Operation represents a protocol operation that can be halted.
type Operation interface {
	baseOp
	Type() OpType
}

// OpType represents the type of operation.
type OpType string

// Predefined operation types in the protocol.
const (
	OpTypePool          OpType = rawPool
	OpTypePosition      OpType = rawPosition
	OpTypeProtocolFee   OpType = rawProtocolFee
	OpTypeRouter        OpType = rawRouter
	OpTypeStaker        OpType = rawStaker
	OpTypeLaunchpad     OpType = rawLaunchpad
	OpTypeGovernance    OpType = rawGovernance
	OpTypeGovStaker     OpType = rawGovStaker
	OpTypeXGns          OpType = rawXGns
	OpTypeCommunityPool OpType = rawCommunityPool
	OpTypeEmission      OpType = rawEmission
	OpTypeWithdraw      OpType = rawWithdraw
)

func (o OpType) String() string { return string(o) }

// basicOperation implements the `Operation` interface.
type basicOperation struct {
	baseInfo
	opType OpType
}

func (o basicOperation) Type() OpType { return o.opType }

// NewOperation creates a new `BasicOperation` instance.
func NewOperation(opType OpType, name, desc string) Operation {
	return basicOperation{newBaseInfo(name, desc), opType}
}
