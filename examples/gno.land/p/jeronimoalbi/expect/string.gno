package expect

import (
	"std"
	"strings"
)

type StringChecker interface {
	Not() StringChecker
	ToEqual(string)
	ToBeEmpty()
	ToHaveLength(int)
}

type stringChecker struct {
	ctx   context
	value string
}

func (c stringChecker) Not() StringChecker {
	c.ctx.not = !c.ctx.not
	return c
}

func (c stringChecker) ToEqual(v string) {
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

func (c stringChecker) ToBeEmpty() {
	c.ctx.t.Helper()

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

func (c stringChecker) ToHaveLength(length int) {
	c.ctx.t.Helper()

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
