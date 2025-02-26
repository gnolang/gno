package halt

var _ Operation = (*basicOperation)(nil)

const (
	rawSwap       = "swap"
	rawLiquidity  = "liquidity"
	rawWithdraw   = "withdraw"
	rawGovernance = "governance"
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
	OpTypeSwap       OpType = rawSwap
	OpTypeLiquidity  OpType = rawLiquidity
	OpTypeWithdraw   OpType = rawWithdraw
	OpTypeGovernance OpType = rawGovernance
)

func (o OpType) String() string { return string(o) }

// OpTypeFromString converts a string to an `OpType`.
func OpTypeFromString(s string) (OpType, error) {
	switch s {
	case rawSwap:
		return OpTypeSwap, nil
	case rawLiquidity:
		return OpTypeLiquidity, nil
	case rawWithdraw:
		return OpTypeWithdraw, nil
	case rawGovernance:
		return OpTypeGovernance, nil
	default:
		// handle user-defined operation types
		return OpType(s), nil
	}
}

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
