package expect

import (
	"strconv"

	"gno.land/p/demo/ufmt"
)

// NewFloatChecker creates a new checker of float64 values.
func NewFloatChecker(ctx Context, value float64) FloatChecker {
	return FloatChecker{ctx, value}
}

// FloatChecker asserts float64 values.
type FloatChecker struct {
	ctx   Context
	value float64
}

// Not negates the next called expectation.
func (c FloatChecker) Not() FloatChecker {
	c.ctx.negated = !c.ctx.negated
	return c
}

// ToEqual asserts that current value is equal to an expected value.
func (c FloatChecker) ToEqual(value float64) {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.value == value, func(ctx Context) string {
		if !ctx.IsNegated() {
			got := c.format(c.value)
			return ufmt.Sprintf("Expected values to match\nGot: %s\nWant: %s", got, c.format(value))
		}
		return ufmt.Sprintf("Expected value to be different\nGot: %s", c.format(c.value))
	})
}

// ToBeGreaterThan asserts that current value is greater than an expected value.
func (c FloatChecker) ToBeGreaterThan(value float64) {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.value > value, func(ctx Context) string {
		got := c.format(c.value)
		if !ctx.IsNegated() {
			return ufmt.Sprintf("Expected values to be gerater than %s\nGot: %s", c.format(value), got)
		}
		return ufmt.Sprintf("Expected value to not to be greater than %s\nGot: %s", c.format(value), got)
	})
}

// ToBeGreaterOrEqualThan asserts that current value is greater or equal than an expected value.
func (c FloatChecker) ToBeGreaterOrEqualThan(value float64) {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.value >= value, func(ctx Context) string {
		got := c.format(c.value)
		if !ctx.IsNegated() {
			return ufmt.Sprintf("Expected values to be greater or equal than %s\nGot: %s", c.format(value), got)
		}
		return ufmt.Sprintf("Expected value to not to be greater or equal than %s\nGot: %s", c.format(value), got)
	})
}

// ToBeLowerThan asserts that current value is lower than an expected value.
func (c FloatChecker) ToBeLowerThan(value float64) {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.value < value, func(ctx Context) string {
		got := c.format(c.value)
		if !ctx.IsNegated() {
			return ufmt.Sprintf("Expected values to be lower than %s\nGot: %s", c.format(value), got)
		}
		return ufmt.Sprintf("Expected value to not to be lower than %s\nGot: %s", c.format(value), got)
	})
}

// ToBeLowerOrEqualThan asserts that current value is lower or equal than an expected value.
func (c FloatChecker) ToBeLowerOrEqualThan(value float64) {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.value <= value, func(ctx Context) string {
		got := c.format(c.value)
		if !ctx.IsNegated() {
			return ufmt.Sprintf("Expected values to be lower or equal than %s\nGot: %s", c.format(value), got)
		}
		return ufmt.Sprintf("Expected value to not to be lower or equal than %s\nGot: %s", c.format(value), got)
	})
}

func (FloatChecker) format(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}

func asFloat(value any) (float64, error) {
	switch v := value.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, ErrIncompatibleType
	}
}
