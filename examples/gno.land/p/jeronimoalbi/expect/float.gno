package expect

import "strconv"

func NewFloatChecker(ctx Context, value float64) FloatChecker {
	return FloatChecker{ctx, value}
}

type FloatChecker struct {
	ctx   Context
	value float64
}

func (c FloatChecker) Not() FloatChecker {
	c.ctx.not = !c.ctx.not
	return c
}

func (c FloatChecker) ToEqual(value float64) {
	c.ctx.t().Helper()
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

func (c FloatChecker) ToBeGreaterThan(value float64) {
	c.ctx.t().Helper()
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

func (c FloatChecker) ToBeGreaterOrEqualThan(value float64) {
	c.ctx.t().Helper()
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

func (c FloatChecker) ToBeLowerThan(value float64) {
	c.ctx.t().Helper()
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

func (c FloatChecker) ToBeLowerOrEqualThan(value float64) {
	c.ctx.t().Helper()
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
