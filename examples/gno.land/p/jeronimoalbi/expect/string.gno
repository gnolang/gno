package expect

import (
	"errors"

	"gno.land/p/nt/ufmt"
)

// ErrIncompatibleType indicates that a value can't be casted to a different type.
var ErrIncompatibleType = errors.New("incompatible type")

// NewStringChecker creates a new checker of string values.
func NewStringChecker(ctx Context, value string) StringChecker {
	return StringChecker{ctx, value}
}

// StringChecker asserts string values.
type StringChecker struct {
	ctx   Context
	value string
}

// Not negates the next called expectation.
func (c StringChecker) Not() StringChecker {
	c.ctx.negated = !c.ctx.negated
	return c
}

// ToEqual asserts that current value is equal to an expected value.
func (c StringChecker) ToEqual(v string) {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.value == v, func(ctx Context) string {
		if !ctx.IsNegated() {
			return ufmt.Sprintf("Expected values to match\nGot: %s\nWant: %s", c.value, v)
		}
		return ufmt.Sprintf("Expected values to be different\nGot: %s", c.value)
	})
}

// ToBeEmpty asserts that current value is an empty string.
func (c StringChecker) ToBeEmpty() {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.value == "", func(ctx Context) string {
		if !ctx.IsNegated() {
			return ufmt.Sprintf("Expected string to be empty\nGot: %s", c.value)
		}
		return ufmt.Sprintf("Unexpected empty string")
	})
}

// ToHaveLength asserts that current value has an expected length.
func (c StringChecker) ToHaveLength(length int) {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(len(c.value) == length, func(ctx Context) string {
		got := len(c.value)
		if !ctx.IsNegated() {
			return ufmt.Sprintf("Expected string length to match\nGot: %d\nWant: %d", got, length)
		}
		return ufmt.Sprintf("Expected string lengths to be different\nGot: %d", got)
	})
}

// Stringer defines an interface for values that has a String method.
type Stringer interface {
	String() string
}

func asString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case Stringer:
		return v.String(), nil
	case address:
		return v.String(), nil
	default:
		return "", ErrIncompatibleType
	}
}
