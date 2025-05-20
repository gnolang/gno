package expect

import "strconv"

func NewUintChecker(ctx Context, value uint64) UintChecker {
	return UintChecker{ctx, value}
}

type UintChecker struct {
	ctx   Context
	value uint64
}

func (c UintChecker) Not() UintChecker {
	c.ctx.not = !c.ctx.not
	return c
}

func (c UintChecker) ToEqual(value uint64) {
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

func (c UintChecker) ToBeGreaterThan(value uint64) {
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

func (c UintChecker) ToBeGreaterOrEqualThan(value uint64) {
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

func (c UintChecker) ToBeLowerThan(value uint64) {
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

func (c UintChecker) ToBeLowerOrEqualThan(value uint64) {
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

func (UintChecker) format(value uint64) string {
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
