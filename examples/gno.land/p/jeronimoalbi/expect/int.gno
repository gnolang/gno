package expect

import "strconv"

type IntChecker interface {
	Not() IntChecker
	ToEqual(int64)
	ToBeGreaterThan(int64)
	ToBeGreaterOrEqualThan(int64)
	ToBeLowerThan(int64)
	ToBeLowerOrEqualThan(int64)
}

type intChecker struct {
	ctx   context
	value int64
}

func (c intChecker) Not() IntChecker {
	c.ctx.not = !c.ctx.not
	return c
}

func (c intChecker) ToEqual(value int64) {
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

func (c intChecker) ToBeGreaterThan(value int64) {
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

func (c intChecker) ToBeGreaterOrEqualThan(value int64) {
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

func (c intChecker) ToBeLowerThan(value int64) {
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

func (c intChecker) ToBeLowerOrEqualThan(value int64) {
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

func (intChecker) format(value int64) string {
	return strconv.FormatInt(value, 10)
}

func asInt(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	default:
		return 0, ErrIncompatibleType
	}
}
