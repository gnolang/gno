package expect

import (
	"strconv"

	"gno.land/p/nt/ufmt/v0"
)

// NewBooleanChecker creates a new checker of boolean values
func NewBooleanChecker(ctx Context, value bool) BooleanChecker {
	return BooleanChecker{ctx, value}
}

// BooleanChecker asserts boolean values.
type BooleanChecker struct {
	ctx   Context
	value bool
}

// Not negates the next called expectation.
func (c BooleanChecker) Not() BooleanChecker {
	c.ctx.negated = !c.ctx.negated
	return c
}

// ToEqual asserts that current value is equal to an expected value.
func (c BooleanChecker) ToEqual(v bool) {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.value == v, func(ctx Context) string {
		got := formatBoolean(c.value)
		if !ctx.IsNegated() {
			want := formatBoolean(v)
			return ufmt.Sprintf("Expected values to match\nGot: %s\nWant: %s", got, want)
		}
		return ufmt.Sprintf("Expected values to be different\nGot: %s", got)
	})
}

// ToBeFalsy asserts that current value is falsy.
func (c BooleanChecker) ToBeFalsy() {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(!c.value, func(ctx Context) string {
		if !ctx.IsNegated() {
			return "Expected value to be falsy"
		}
		return "Expected value not to be falsy"
	})
}

// ToBeTruthy asserts that current value is truthy.
func (c BooleanChecker) ToBeTruthy() {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.value, func(ctx Context) string {
		if !ctx.IsNegated() {
			return "Expected value to be truthy"
		}
		return "Expected value not to be truthy"
	})
}

func asBoolean(value any) (bool, error) {
	if value == nil {
		return false, nil
	}

	var s string
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		s = v
	case []byte:
		s = string(v)
	case Stringer:
		s = v.String()
	default:
		return false, ErrIncompatibleType
	}

	if s != "" {
		return strconv.ParseBool(s)
	}
	return false, nil
}

func formatBoolean(value bool) string {
	return strconv.FormatBool(value)
}
