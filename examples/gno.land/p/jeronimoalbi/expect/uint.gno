package expect

import "strconv"

type UintChecker interface {
	Not() UintChecker
	ToEqual(uint64)
	ToBeGreaterThan(uint64)
	ToBeGreaterOrEqualThan(uint64)
	ToBeLowerThan(uint64)
	ToBeLowerOrEqualThan(uint64)
}

type uintChecker struct {
	ctx   context
	value uint64
}

func (c uintChecker) Not() UintChecker {
	c.ctx.not = !c.ctx.not
	return c
}

func (c uintChecker) ToEqual(value uint64) {
	c.ctx.t.Helper()

	c.ctx.CheckExpectation(
		c.value == value,
		func() {
			c.ctx.Fail("Expected value to be different\nGot: %s", c.format(c.value))
		},
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected values to match\nGot: %s\nWant: %s", got, c.format(value))
		},
	)
}

func (c uintChecker) ToBeGreaterThan(value uint64) {
	c.ctx.t.Helper()

	c.ctx.CheckExpectation(
		c.value > value,
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected value to not to be greater than %s\nGot: %s", c.format(value), got)
		},
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected values to be gerater than %s\nGot: %s", c.format(value), got)
		},
	)
}

func (c uintChecker) ToBeGreaterOrEqualThan(value uint64) {
	c.ctx.t.Helper()

	c.ctx.CheckExpectation(
		c.value >= value,
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected value to not to be greater or equal than %s\nGot: %s", c.format(value), got)
		},
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected values to be greater or equal than %s\nGot: %s", c.format(value), got)
		},
	)
}

func (c uintChecker) ToBeLowerThan(value uint64) {
	c.ctx.t.Helper()

	c.ctx.CheckExpectation(
		c.value < value,
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected value to not to be lower than %s\nGot: %s", c.format(value), got)
		},
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected values to be lower than %s\nGot: %s", c.format(value), got)
		},
	)
}

func (c uintChecker) ToBeLowerOrEqualThan(value uint64) {
	c.ctx.t.Helper()

	c.ctx.CheckExpectation(
		c.value <= value,
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected value to not to be lower or equal than %s\nGot: %s", c.format(value), got)
		},
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected values to be lower or equal than %s\nGot: %s", c.format(value), got)
		},
	)
}

func (uintChecker) format(value uint64) string {
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
