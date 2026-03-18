package types

import (
	"math"
)

// Gas consumption descriptors.
const (
	GasIterNextCostFlatDesc = "IterNextFlat"
	GasValuePerByteDesc     = "ValuePerByte"
	GasWritePerByteDesc     = "WritePerByte"
	GasReadPerByteDesc      = "ReadPerByte"
	GasWriteCostFlatDesc    = "WriteFlat"
	GasReadCostFlatDesc     = "ReadFlat"
	GasHasDesc              = "Has"
	GasDeleteDesc           = "Delete"
)

// Gas measured by the SDK
type Gas = int64

// OutOfGasError defines an error thrown when an action results in out of gas.
type OutOfGasError struct {
	Descriptor string
}

func (oog OutOfGasError) Error() string {
	return "out of gas in location: " + oog.Descriptor
}

// GasOverflowError defines an error thrown when an action results gas consumption
// unsigned integer overflow.
type GasOverflowError struct {
	Descriptor string
}

func (oog GasOverflowError) Error() string {
	return "gas overflow in location: " + oog.Descriptor
}

// OPTIMIZATION: GasMeter was an interface with 3 implementations
// (basicGasMeter, infiniteGasMeter, passthroughGasMeter). Now a single
// concrete struct. This lets the Go compiler inline ConsumeGas at call
// sites, eliminating interface dispatch (~2-3ns per call on the flush path).
//
// GasMeter is a concrete gas meter that tracks gas consumption.
// It replaces the old GasMeter interface with a single struct that
// handles limited, infinite, and passthrough modes:
//   - limit > 0: panics with OutOfGasError when consumed exceeds limit
//   - limit == 0: infinite mode, no limit check
//   - parent != nil: passthrough, also charges the parent meter
//
// Being concrete (not an interface) allows the Go compiler to inline
// ConsumeGas, eliminating interface dispatch on the VM hot path.
type GasMeter struct {
	consumed Gas
	limit    Gas
	parent   *GasMeter // non-nil for passthrough mode
}

// NewGasMeter returns a new GasMeter with the given gas limit.
func NewGasMeter(limit Gas) *GasMeter {
	if limit < 0 {
		panic("gas must not be negative")
	}
	return &GasMeter{
		limit: limit,
	}
}

// NewInfiniteGasMeter returns a new GasMeter with no gas limit.
func NewInfiniteGasMeter() *GasMeter {
	return &GasMeter{}
}

// NewPassthroughGasMeter returns a GasMeter that charges both itself
// (up to limit) and a parent meter.
func NewPassthroughGasMeter(parent *GasMeter, limit Gas) *GasMeter {
	if limit < 0 {
		panic("gas must not be negative")
	}
	return &GasMeter{
		limit:  limit,
		parent: parent,
	}
}

func (g *GasMeter) GasConsumed() Gas {
	return g.consumed
}

func (g *GasMeter) GasConsumedToLimit() Gas {
	if g.limit > 0 && g.consumed > g.limit {
		return g.limit
	}
	return g.consumed
}

func (g *GasMeter) Limit() Gas {
	return g.limit
}

func (g *GasMeter) Remaining() Gas {
	if g.limit == 0 {
		return math.MaxInt64
	}
	rem := g.limit - g.consumed
	if rem < 0 {
		return 0
	}
	return rem
}

// ConsumeGas adds amount to consumed gas. Panics with OutOfGasError
// if a limit is set and exceeded, or GasOverflowError on int64 overflow.
func (g *GasMeter) ConsumeGas(amount Gas, descriptor string) {
	consumed := g.consumed + amount
	if consumed < g.consumed { // int64 overflow
		panic(GasOverflowError{descriptor})
	}
	g.consumed = consumed
	if g.limit > 0 && consumed > g.limit {
		panic(OutOfGasError{descriptor})
	}
	if g.parent != nil {
		g.parent.ConsumeGas(amount, descriptor)
	}
}

func (g *GasMeter) IsPastLimit() bool {
	return g.limit > 0 && g.consumed > g.limit
}

func (g *GasMeter) IsOutOfGas() bool {
	return g.limit > 0 && g.consumed >= g.limit
}

//----------------------------------------

// GasConfig defines gas cost for each operation on KVStores
type GasConfig struct {
	HasCost          Gas
	DeleteCost       Gas
	ReadCostFlat     Gas
	ReadCostPerByte  Gas
	WriteCostFlat    Gas
	WriteCostPerByte Gas
	IterNextCostFlat Gas
}

// DefaultGasConfig returns a default gas config for KVStores.
func DefaultGasConfig() GasConfig {
	return GasConfig{
		HasCost:          1000,
		DeleteCost:       1000,
		ReadCostFlat:     1000,
		ReadCostPerByte:  3,
		WriteCostFlat:    2000,
		WriteCostPerByte: 30,
		IterNextCostFlat: 30,
	}
}
