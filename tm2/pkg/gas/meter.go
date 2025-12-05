package gas

import (
	"math"

	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// Meter interface to track gas consumption.
type Meter interface {
	GasConsumed() Gas
	GasConsumedToLimit() Gas
	Limit() Gas
	Remaining() Gas
	ConsumeGas(amount Gas, descriptor string)
	IsPastLimit() bool
	IsOutOfGas() bool
}

//----------------------------------------
// basicMeter

type basicMeter struct {
	limit    Gas
	consumed Gas
}

// NewMeter returns a reference to a new basicMeter.
func NewMeter(limit Gas) *basicMeter {
	if limit < 0 {
		panic("gas must not be negative")
	}
	return &basicMeter{
		limit:    limit,
		consumed: 0,
	}
}

func (g *basicMeter) GasConsumed() Gas {
	return g.consumed
}

func (g *basicMeter) Limit() Gas {
	return g.limit
}

func (g *basicMeter) Remaining() Gas {
	return overflow.Subp(g.Limit(), g.GasConsumedToLimit())
}

func (g *basicMeter) GasConsumedToLimit() Gas {
	if g.IsPastLimit() {
		return g.limit
	}
	return g.consumed
}

// TODO rename to DidConsumeGas.
func (g *basicMeter) ConsumeGas(amount Gas, descriptor string) {
	if amount < 0 {
		panic("gas must not be negative")
	}
	consumed, ok := overflow.Add(g.consumed, amount)
	if !ok {
		panic(OverflowError{descriptor})
	}
	// consume gas even if out of gas.
	// corollary, call (Did)ConsumeGas after consumption.
	g.consumed = consumed
	if consumed > g.limit {
		panic(OutOfGasError{descriptor})
	}
}

func (g *basicMeter) IsPastLimit() bool {
	return g.consumed > g.limit
}

func (g *basicMeter) IsOutOfGas() bool {
	return g.consumed >= g.limit
}

//----------------------------------------
// infiniteMeter

type infiniteMeter struct {
	consumed Gas
}

// NewInfiniteMeter returns a reference to a new infiniteMeter.
func NewInfiniteMeter() Meter {
	return &infiniteMeter{
		consumed: 0,
	}
}

func (g *infiniteMeter) GasConsumed() Gas {
	return g.consumed
}

func (g *infiniteMeter) GasConsumedToLimit() Gas {
	return g.consumed
}

func (g *infiniteMeter) Limit() Gas {
	return 0
}

func (g *infiniteMeter) Remaining() Gas {
	return math.MaxInt64
}

func (g *infiniteMeter) ConsumeGas(amount Gas, descriptor string) {
	consumed, ok := overflow.Add(g.consumed, amount)
	if !ok {
		panic(OverflowError{descriptor})
	}
	g.consumed = consumed
}

func (g *infiniteMeter) IsPastLimit() bool {
	return false
}

func (g *infiniteMeter) IsOutOfGas() bool {
	return false
}

//----------------------------------------
// passthroughMeter

type passthroughMeter struct {
	Base Meter
	Head *basicMeter
}

// NewPassthroughMeter has a head basicMeter, but also passes through
// consumption to a base basicMeter.  Limit must be less than
// base.Remaining().
func NewPassthroughMeter(base Meter, limit int64) passthroughMeter {
	if limit < 0 {
		panic("gas must not be negative")
	}
	// limit > base.Remaining() is not checked; so that a panic happens when
	// gas is actually consumed.
	return passthroughMeter{
		Base: base,
		Head: NewMeter(limit),
	}
}

func (g passthroughMeter) GasConsumed() Gas {
	return g.Head.GasConsumed()
}

func (g passthroughMeter) Limit() Gas {
	return g.Head.Limit()
}

func (g passthroughMeter) Remaining() Gas {
	return g.Head.Remaining()
}

func (g passthroughMeter) GasConsumedToLimit() Gas {
	return g.Head.GasConsumedToLimit()
}

func (g passthroughMeter) ConsumeGas(amount Gas, descriptor string) {
	g.Base.ConsumeGas(amount, descriptor)
	g.Head.ConsumeGas(amount, descriptor)
}

func (g passthroughMeter) IsPastLimit() bool {
	return g.Head.IsPastLimit()
}

func (g passthroughMeter) IsOutOfGas() bool {
	return g.Head.IsOutOfGas()
}
