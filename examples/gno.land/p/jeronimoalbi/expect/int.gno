package expect

import "strconv"

// NewIntChecker creates a new checker of int64 values.
func NewIntChecker(ctx Context, value int64) IntChecker {
	return IntChecker{ctx, value}
}

// IntChecker asserts int64 values.
type IntChecker struct {
	ctx   Context
	value int64
}

// Not negates the next called expectation.
func (c IntChecker) Not() IntChecker {
	c.ctx.negated = !c.ctx.negated
	return c
}

// ToEqual asserts that current value is equal to an expected value.
func (c IntChecker) ToEqual(value int64) {
	c.ctx.t().Helper()
	c.ctx.CheckExpectation(
		c.value == value,
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected values to match\nGot: %s\nWant: %s", got, c.format(value))
		},
		func() {
			c.ctx.Fail("Expected value to be different\nGot: %s", c.format(c.value))
		},
	)
}

// ToBeGreaterThan asserts that current value is greater than an expected value.
func (c IntChecker) ToBeGreaterThan(value int64) {
	c.ctx.t().Helper()
	c.ctx.CheckExpectation(
		c.value > value,
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected values to be gerater than %s\nGot: %s", c.format(value), got)
		},
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected value to not to be greater than %s\nGot: %s", c.format(value), got)
		},
	)
}

// ToBeGreaterOrEqualThan asserts that current value is greater or equal than an expected value.
func (c IntChecker) ToBeGreaterOrEqualThan(value int64) {
	c.ctx.t().Helper()
	c.ctx.CheckExpectation(
		c.value >= value,
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected values to be greater or equal than %s\nGot: %s", c.format(value), got)
		},
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected value to not to be greater or equal than %s\nGot: %s", c.format(value), got)
		},
	)
}

// ToBeLowerThan asserts that current value is lower than an expected value.
func (c IntChecker) ToBeLowerThan(value int64) {
	c.ctx.t().Helper()
	c.ctx.CheckExpectation(
		c.value < value,
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected values to be lower than %s\nGot: %s", c.format(value), got)
		},
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected value to not to be lower than %s\nGot: %s", c.format(value), got)
		},
	)
}

// ToBeLowerOrEqualThan asserts that current value is lower or equal than an expected value.
func (c IntChecker) ToBeLowerOrEqualThan(value int64) {
	c.ctx.t().Helper()
	c.ctx.CheckExpectation(
		c.value <= value,
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected values to be lower or equal than %s\nGot: %s", c.format(value), got)
		},
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected value to not to be lower or equal than %s\nGot: %s", c.format(value), got)
		},
	)
}

func (IntChecker) format(value int64) string {
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
