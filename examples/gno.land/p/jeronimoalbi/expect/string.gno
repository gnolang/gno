package expect

import (
	"errors"
	"std"
	"strings"
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
	c.ctx.t().Helper()
	c.ctx.CheckExpectation(
		c.value == v,
		func() {
			c.ctx.Fail("Expected values to match\nGot: %s\nWant: %s", c.value, v)
		},
		func() {
			c.ctx.Fail("Expected values to be different\nGot: %s", c.value)
		},
	)
}

// ToBeEmpty asserts that current value is an empty string.
func (c StringChecker) ToBeEmpty() {
	c.ctx.t().Helper()

	got := strings.TrimSpace(c.value)
	c.ctx.CheckExpectation(
		got == "",
		func() {
			c.ctx.Fail("Expected string to be empty\nGot: %s", got)
		},
		func() {
			c.ctx.Fail("Unexpected empty string")
		},
	)
}

// ToHaveLength asserts that current value has an expected length.
func (c StringChecker) ToHaveLength(length int) {
	c.ctx.t().Helper()

	got := len(c.value)
	c.ctx.CheckExpectation(
		got == length,
		func() {
			c.ctx.Fail("Expected string length to match\nGot: %d\nWant: %d", got, length)
		},
		func() {
			c.ctx.Fail("Expected string lengths to be different\nGot: %d", got)
		},
	)
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
	case std.Address:
		return v.String(), nil
	default:
		return "", ErrIncompatibleType
	}
}
