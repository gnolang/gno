package expect

import (
	"strconv"

	"gno.land/p/nt/ufmt"
)

// NewUintChecker creates a new checker of uint64 values.
func NewUintChecker(ctx Context, value uint64) UintChecker {
	return UintChecker{ctx, value}
}

// UintChecker asserts uint64 values.
type UintChecker struct {
	ctx   Context
	value uint64
}

// Not negates the next called expectation.
func (c UintChecker) Not() UintChecker {
	c.ctx.negated = !c.ctx.negated
	return c
}

// ToEqual asserts that current value is equal to an expected value.
func (c UintChecker) ToEqual(value uint64) {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.value == value, func(ctx Context) string {
		if !ctx.IsNegated() {
			got := formatUint(c.value)
			want := formatUint(value)
			return ufmt.Sprintf("Expected values to match\nGot: %s\nWant: %s", got, want)
		}
		return ufmt.Sprintf("Expected value to be different\nGot: %s", formatUint(c.value))
	})
}

// ToBeGreaterThan asserts that current value is greater than an expected value.
func (c UintChecker) ToBeGreaterThan(value uint64) {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.value > value, func(ctx Context) string {
		got := formatUint(c.value)
		want := formatUint(value)
		if !ctx.IsNegated() {
			return ufmt.Sprintf("Expected values to be gerater than %s\nGot: %s", want, got)
		}
		return ufmt.Sprintf("Expected value to not to be greater than %s\nGot: %s", want, got)
	})
}

// ToBeGreaterOrEqualThan asserts that current value is greater or equal than an expected value.
func (c UintChecker) ToBeGreaterOrEqualThan(value uint64) {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.value >= value, func(ctx Context) string {
		got := formatUint(c.value)
		want := formatUint(value)
		if !ctx.IsNegated() {
			return ufmt.Sprintf("Expected values to be greater or equal than %s\nGot: %s", want, got)
		}
		return ufmt.Sprintf("Expected value to not to be greater or equal than %s\nGot: %s", want, got)
	})
}

// ToBeLowerThan asserts that current value is lower than an expected value.
func (c UintChecker) ToBeLowerThan(value uint64) {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.value < value, func(ctx Context) string {
		got := formatUint(c.value)
		want := formatUint(value)
		if !ctx.IsNegated() {
			return ufmt.Sprintf("Expected values to be lower than %s\nGot: %s", want, got)
		}
		return ufmt.Sprintf("Expected value to not to be lower than %s\nGot: %s", want, got)
	})
}

// ToBeLowerOrEqualThan asserts that current value is lower or equal than an expected value.
func (c UintChecker) ToBeLowerOrEqualThan(value uint64) {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.value <= value, func(ctx Context) string {
		got := formatUint(c.value)
		want := formatUint(value)
		if !ctx.IsNegated() {
			return ufmt.Sprintf("Expected values to be lower or equal than %s\nGot: %s", want, got)
		}
		return ufmt.Sprintf("Expected value to not to be lower or equal than %s\nGot: %s", want, got)
	})
}

func formatUint(value uint64) string {
	return strconv.FormatUint(value, 10)
}

func asUint(value any) (uint64, error) {
	switch v := value.(type) {
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case int:
		if v < 0 {
			return 0, ErrIncompatibleType
		}
		return uint64(v), nil
	default:
		return 0, ErrIncompatibleType
	}
}
