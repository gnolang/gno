package expect

import "strconv"

func NewBooleanChecker(ctx Context, value bool) BooleanChecker {
	return BooleanChecker{ctx, value}
}

type BooleanChecker struct {
	ctx   Context
	value bool
}

func (c BooleanChecker) Not() BooleanChecker {
	c.ctx.not = !c.ctx.not
	return c
}

func (c BooleanChecker) ToEqual(v bool) {
	c.ctx.t().Helper()
	c.ctx.CheckExpectation(
		c.value == v,
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected values to be different\nGot: %s", got)
		},
		func() {
			got := c.format(c.value)
			c.ctx.Fail("Expected values to match\nGot: %s\nWant: %s", got, strconv.FormatBool(v))
		},
	)
}

func (c BooleanChecker) ToBeFalsy() {
	c.ctx.t().Helper()
	c.ctx.CheckExpectation(
		!c.value,
		func() {
			c.ctx.Fail("Expected value not to be falsy")
		},
		func() {
			c.ctx.Fail("Expected value to be falsy")
		},
	)
}

func (c BooleanChecker) ToBeTruthy() {
	c.ctx.t().Helper()
	c.ctx.CheckExpectation(
		c.value,
		func() {
			c.ctx.Fail("Expected value not to be truthy")
		},
		func() {
			c.ctx.Fail("Expected value to be truthy")
		},
	)
}

func (BooleanChecker) format(value bool) string {
	return strconv.FormatBool(value)
}

func asBoolean(value any) (bool, error) {
	if value == nil {
		return false, nil
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		if v == "" {
			return false, nil
		}
		return strconv.ParseBool(v)
	case []byte:
		return v != nil, nil
	case Stringer:
		s := v.String()
		if s == "" {
			return false, nil
		}
		return strconv.ParseBool(v.String())
	default:
		return false, ErrIncompatibleType
	}
}
