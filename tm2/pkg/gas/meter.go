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
	ConsumeGas(operation Operation, multiplier float64)
	CalculateGasCost(operation Operation, multiplier float64) Gas
	IsPastLimit() bool
	IsOutOfGas() bool
	Config() Config
}

// calculateGasCost calculates the gas cost for a given operation, multiplier
// and global multiplier from the config.
func calculateGasCost(config *Config, operation Operation, multiplier float64) Gas {
	// Get the operation cost from the config.
	operationCost := config.Costs[operation]

	// Calculate base cost with multiplier.
	basecost, ok := overflow.Mul(float64(operationCost), multiplier)
	if !ok {
		panic(OverflowError{operation.String()})
	}

	// Calculate total cost with global multiplier.
	totalCost, ok := overflow.Mul(basecost, config.GlobalMultiplier)
	if !ok {
		panic(OverflowError{operation.String()})
	}

	// Round to the nearest whole number if there's any fractional part.
	roundedCost := math.Round(totalCost)

	return Gas(roundedCost)
}

//----------------------------------------
// basicMeter

type basicMeter struct {
	limit    Gas
	consumed Gas
	config   Config
}

// NewMeter returns a reference to a new basicMeter with the provided configuration.
func NewMeter(limit Gas, config Config) *basicMeter {
	if limit < 0 {
		panic("gas must not be negative")
	}
	if config.GlobalMultiplier <= 0 {
		panic("config multiplier must be positive")
	}
	return &basicMeter{
		limit:    limit,
		consumed: 0,
		config:   config,
	}
}

func (g *basicMeter) GasConsumed() Gas {
	return g.consumed
}

func (g *basicMeter) GasConsumedToLimit() Gas {
	if g.IsPastLimit() {
		return g.limit
	}
	return g.consumed
}

func (g *basicMeter) Limit() Gas {
	return g.limit
}

func (g *basicMeter) Remaining() Gas {
	return overflow.Subp(g.Limit(), g.GasConsumedToLimit())
}

func (g *basicMeter) ConsumeGas(operation Operation, multiplier float64) {
	gasCost := g.CalculateGasCost(operation, multiplier)

	consumed, ok := overflow.Add(g.consumed, gasCost)
	if !ok {
		panic(OverflowError{operation.String()})
	}
	// Consume gas even if out of gas.
	// Corollary, call ConsumeGas after consumption.
	g.consumed = consumed
	if consumed > g.limit {
		panic(OutOfGasError{operation.String()})
	}
}

func (g *basicMeter) CalculateGasCost(operation Operation, multiplier float64) Gas {
	return calculateGasCost(&g.config, operation, multiplier)
}

func (g *basicMeter) IsPastLimit() bool {
	return g.consumed > g.limit
}

func (g *basicMeter) IsOutOfGas() bool {
	return g.consumed >= g.limit
}

func (g *basicMeter) Config() Config {
	return g.config
}

//----------------------------------------
// infiniteMeter

type infiniteMeter struct {
	consumed Gas
	config   Config
}

// NewInfiniteMeter returns a reference to a new infiniteMeter with the provided configuration.
func NewInfiniteMeter(config Config) Meter {
	if config.GlobalMultiplier <= 0 {
		panic("config multiplier must be positive")
	}
	return &infiniteMeter{
		consumed: 0,
		config:   config,
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

func (g *infiniteMeter) ConsumeGas(operation Operation, multiplier float64) {
	gasCost := g.CalculateGasCost(operation, multiplier)

	consumed, ok := overflow.Add(g.consumed, gasCost)
	if !ok {
		panic(OverflowError{operation.String()})
	}
	g.consumed = consumed
}

func (g *infiniteMeter) CalculateGasCost(operation Operation, multiplier float64) Gas {
	return calculateGasCost(&g.config, operation, multiplier)
}

func (g *infiniteMeter) IsPastLimit() bool {
	return false
}

func (g *infiniteMeter) IsOutOfGas() bool {
	return false
}

func (g *infiniteMeter) Config() Config {
	return g.config
}

//----------------------------------------
// passthroughMeter

type passthroughMeter struct {
	Base Meter
	Head *basicMeter
}

// NewPassthroughMeter has a head basicMeter, but also passes through
// consumption to a base basicMeter. Limit must be less than
// base.Remaining().
func NewPassthroughMeter(base Meter, limit int64, config Config) passthroughMeter {
	if limit < 0 {
		panic("gas must not be negative")
	}
	// limit > base.Remaining() is not checked; so that a panic happens when
	// gas is actually consumed.
	return passthroughMeter{
		Base: base,
		Head: NewMeter(limit, config),
	}
}

func (g passthroughMeter) GasConsumed() Gas {
	return g.Head.GasConsumed()
}

func (g passthroughMeter) GasConsumedToLimit() Gas {
	return g.Head.GasConsumedToLimit()
}

func (g passthroughMeter) Limit() Gas {
	return g.Head.Limit()
}

func (g passthroughMeter) Remaining() Gas {
	return g.Head.Remaining()
}

func (g passthroughMeter) ConsumeGas(operation Operation, multiplier float64) {
	g.Base.ConsumeGas(operation, multiplier)
	g.Head.ConsumeGas(operation, multiplier)
}

func (g passthroughMeter) CalculateGasCost(operation Operation, multiplier float64) Gas {
	return g.Head.CalculateGasCost(operation, multiplier)
}

func (g passthroughMeter) IsPastLimit() bool {
	return g.Head.IsPastLimit()
}

func (g passthroughMeter) IsOutOfGas() bool {
	return g.Head.IsOutOfGas()
}

func (g passthroughMeter) Config() Config {
	return g.Head.Config()
}
