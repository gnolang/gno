package expect

import (
	"errors"
	"std"
	"strings"
)

var ErrIncompatibleType = errors.New("incompatible type")

func NewStringChecker(ctx Context, value string) StringChecker {
	return StringChecker{ctx, value}
}

type StringChecker struct {
	ctx   Context
	value string
}

func (c StringChecker) Not() StringChecker {
	c.ctx.not = !c.ctx.not
	return c
}

func (c StringChecker) ToEqual(v string) {
	c.ctx.t().Helper()
	c.ctx.CheckExpectation(
		c.value == v,
		func() {
			c.ctx.Fail("Expected values to be different\nGot: %s", c.value)
		},
		func() {
			c.ctx.Fail("Expected values to match\nGot: %s\nWant: %s", c.value, v)
		},
	)
}

func (c StringChecker) ToBeEmpty() {
	c.ctx.t().Helper()

	got := strings.TrimSpace(c.value)
	c.ctx.CheckExpectation(
		got == "",
		func() {
			c.ctx.Fail("Unexpected empty string")
		},
		func() {
			c.ctx.Fail("Expected string to be empty\nGot: %s", got)
		},
	)
}

func (c StringChecker) ToHaveLength(length int) {
	c.ctx.t().Helper()

	got := len(c.value)
	c.ctx.CheckExpectation(
		got == length,
		func() {
			c.ctx.Fail("Expected string lengths to be different\nGot: %d", got)
		},
		func() {
			c.ctx.Fail("Expected string length to match\nGot: %d\nWant: %d", got, length)
		},
	)
}

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
