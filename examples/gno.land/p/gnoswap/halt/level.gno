package halt

import (
	"strings"

	"gno.land/p/demo/ufmt"
)

var (
	_ HaltLevel = (*BasicHaltLevel)(nil)
	_ HaltLevel = (*CompositeHaltLevel)(nil)
)

// LevelID represents a unique halt level identifier.
type LevelID uint8

const (
	_               LevelID = iota // undefined
	LvNoHalt                       // normal operation
	LvSwapHalt                     // swap operations halted
	LvLiquidityHalt                // no swaps, no liquidity ops
	LvEmergencyHalt                // only withdrawals allowed
	LvCompleteHalt                 // all ops disabled
)

const (
	rawNoHalt        = "NoHalt"
	rawSwapHalt      = "SwapHalt"
	rawLiquidityHalt = "LiquidityHalt"
	rawEmergencyHalt = "EmergencyHalt"
	rawCompleteHalt  = "CompleteHalt"
	rawUnknown       = "Unknown"
)

// Nth returns the numeric value of the LevelID.
func (lv LevelID) Nth() uint8 { return uint8(lv) }

// String returns the string representation of the LevelID.
func (lv LevelID) String() string {
	switch lv {
	case LvNoHalt:
		return rawNoHalt
	case LvSwapHalt:
		return rawSwapHalt
	case LvLiquidityHalt:
		return rawLiquidityHalt
	case LvEmergencyHalt:
		return rawEmergencyHalt
	case LvCompleteHalt:
		return rawCompleteHalt
	default:
		return rawUnknown
	}
}

// allowedOps is a map that tracks which operation types are allowed.
type allowedOps map[OpType]bool

// HaltLevel represents a protocol halt level with its configuration.
// It defines what operations are allowed at a specific halt level.
type HaltLevel interface {
	baseOp
	Level() LevelID
	IsOperationAllowed(op Operation) bool
}

// HaltLevels represents a collection of halt levels.
// It implements sort.Interface to allow sorting by LevelID.
type HaltLevels []HaltLevel

// Iter iterates through all halt levels and applies the given callback function.
// Returns true if the callback returns true for any iteration, false otherwise.
func (ll HaltLevels) Iter(fn func(HaltLevel) bool) bool {
	for _, lv := range ll {
		if fn(lv) {
			return true
		}
	}
	return false
}

// BasicHaltLevel implements the `HaltLevel` interface.
type BasicHaltLevel struct {
	baseInfo
	level      LevelID
	allowedOps allowedOps
}

func (h BasicHaltLevel) Level() LevelID { return h.level }

// IsOperationAllowed checks if a given operation is allowed at this halt level.
// Returns true if the operation type exists in allowedOps and is marked as allowed.
func (h BasicHaltLevel) IsOperationAllowed(op Operation) bool {
	allowed, exists := h.allowedOps[op.Type()]
	return allowed && exists
}

// NewHaltLevel creates a new HaltLevel instance with the specified configuration.
func NewHaltLevel(level LevelID, name, desc string, allowedOps allowedOps) HaltLevel {
	return BasicHaltLevel{baseInfo{name, desc}, level, allowedOps}
}

// CompositeOp represents the type of operation to be performed
// when combining multiple halt levels.
type CompositeOp uint8

// Available composite operations.
const (
	_              CompositeOp = iota // undefined
	CompositeOpAnd                    // All conditions must be met
	CompositeOpOr                     // At least one condition must be met
)

const (
	rawAnd    = "and"
	rawOr     = "or"
	undefined = "undefined"
)

// CompositeOpFromString converts a string to a CompositeOp.
// Returns an error if the string is not a valid operator.
func CompositeOpFromString(s string) (CompositeOp, error) {
	switch strings.ToLower(s) {
	case rawAnd:
		return CompositeOpAnd, nil
	case rawOr:
		return CompositeOpOr, nil
	default:
		return 0, ufmt.Errorf("unknown composite operator: %s", s)
	}
}

func (o CompositeOp) String() string {
	switch o {
	case CompositeOpAnd:
		return rawAnd
	case CompositeOpOr:
		return rawOr
	default:
		return undefined
	}
}

// CompositeHaltLevel combines multiple halt levels with a logical operator.
type CompositeHaltLevel struct {
	baseInfo
	levels   HaltLevels
	operator CompositeOp
}

func (c *CompositeHaltLevel) Level() LevelID { return 0 /* Placeholder value */ }

// IsOperationAllowed checks if an operation is allowed based on the composite configuration.
//   - AND (CompositeOpAnd): all levels must allow the operation.
//   - OR (CompositeOpOr): at least one level must allow the operation.
func (c *CompositeHaltLevel) IsOperationAllowed(op Operation) bool {
	switch c.operator {
	case CompositeOpAnd:
		return !c.levels.Iter(func(lv HaltLevel) bool {
			return !lv.IsOperationAllowed(op)
		})
	case CompositeOpOr:
		return c.levels.Iter(func(lv HaltLevel) bool {
			return lv.IsOperationAllowed(op)
		})
	default:
		return false
	}
}
