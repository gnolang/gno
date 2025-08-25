package types

import (
	"math"

	"github.com/gnolang/gno/tm2/pkg/overflow"
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

// GasMeter interface to track gas consumption
type GasMeter interface {
	GasConsumed() Gas
	GasConsumedToLimit() Gas
	Limit() Gas
	Remaining() Gas
	ConsumeGas(amount Gas, descriptor string)
	IsPastLimit() bool
	IsOutOfGas() bool
}

//----------------------------------------
// basicGasMeter

type basicGasMeter struct {
	limit    Gas
	consumed Gas
}

// NewGasMeter returns a reference to a new basicGasMeter.
func NewGasMeter(limit Gas) *basicGasMeter {
	if limit < 0 {
		panic("gas must not be negative")
	}
	return &basicGasMeter{
		limit:    limit,
		consumed: 0,
	}
}

func (g *basicGasMeter) GasConsumed() Gas {
	return g.consumed
}

func (g *basicGasMeter) Limit() Gas {
	return g.limit
}

func (g *basicGasMeter) Remaining() Gas {
	return overflow.Subp(g.Limit(), g.GasConsumedToLimit())
}

func (g *basicGasMeter) GasConsumedToLimit() Gas {
	if g.IsPastLimit() {
		return g.limit
	}
	return g.consumed
}

// TODO rename to DidConsumeGas.
func (g *basicGasMeter) ConsumeGas(amount Gas, descriptor string) {
	if amount < 0 {
		panic("gas must not be negative")
	}
	consumed, ok := overflow.Add(g.consumed, amount)
	if !ok {
		panic(GasOverflowError{descriptor})
	}
	// consume gas even if out of gas.
	// corollary, call (Did)ConsumeGas after consumption.
	g.consumed = consumed
	if consumed > g.limit {
		panic(OutOfGasError{descriptor})
	}
}

func (g *basicGasMeter) IsPastLimit() bool {
	return g.consumed > g.limit
}

func (g *basicGasMeter) IsOutOfGas() bool {
	return g.consumed >= g.limit
}

//----------------------------------------
// infiniteGasMeter

type infiniteGasMeter struct {
	consumed Gas
}

// NewInfiniteGasMeter returns a reference to a new infiniteGasMeter.
func NewInfiniteGasMeter() GasMeter {
	return &infiniteGasMeter{
		consumed: 0,
	}
}

func (g *infiniteGasMeter) GasConsumed() Gas {
	return g.consumed
}

func (g *infiniteGasMeter) GasConsumedToLimit() Gas {
	return g.consumed
}

func (g *infiniteGasMeter) Limit() Gas {
	return 0
}

func (g *infiniteGasMeter) Remaining() Gas {
	return math.MaxInt64
}

func (g *infiniteGasMeter) ConsumeGas(amount Gas, descriptor string) {
	consumed, ok := overflow.Add(g.consumed, amount)
	if !ok {
		panic(GasOverflowError{descriptor})
	}
	g.consumed = consumed
}

func (g *infiniteGasMeter) IsPastLimit() bool {
	return false
}

func (g *infiniteGasMeter) IsOutOfGas() bool {
	return false
}

//----------------------------------------
// passthroughGasMeter

type passthroughGasMeter struct {
	Base GasMeter
	Head *basicGasMeter
}

// NewPassthroughGasMeter has a head basicGasMeter, but also passes through
// consumption to a base basicGasMeter.  Limit must be less than
// base.Remaining().
func NewPassthroughGasMeter(base GasMeter, limit int64) passthroughGasMeter {
	if limit < 0 {
		panic("gas must not be negative")
	}
	// limit > base.Remaining() is not checked; so that a panic happens when
	// gas is actually consumed.
	return passthroughGasMeter{
		Base: base,
		Head: NewGasMeter(limit),
	}
}

func (g passthroughGasMeter) GasConsumed() Gas {
	return g.Head.GasConsumed()
}

func (g passthroughGasMeter) Limit() Gas {
	return g.Head.Limit()
}

func (g passthroughGasMeter) Remaining() Gas {
	return g.Head.Remaining()
}

func (g passthroughGasMeter) GasConsumedToLimit() Gas {
	return g.Head.GasConsumedToLimit()
}

func (g passthroughGasMeter) ConsumeGas(amount Gas, descriptor string) {
	g.Base.ConsumeGas(amount, descriptor)
	g.Head.ConsumeGas(amount, descriptor)
}

func (g passthroughGasMeter) IsPastLimit() bool {
	return g.Head.IsPastLimit()
}

func (g passthroughGasMeter) IsOutOfGas() bool {
	return g.Head.IsOutOfGas()
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
