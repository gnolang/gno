package expect

import "strconv"

type BooleanChecker interface {
	ToEqual(bool)
	Not() BooleanChecker
	ToBeFalsy()
	ToBeTruthy()
}

type booleanChecker struct {
	ctx   context
	value bool
}

func (c booleanChecker) Not() BooleanChecker {
	c.ctx.not = !c.ctx.not
	return c
}

func (c booleanChecker) ToEqual(v bool) {
	c.ctx.t.Helper()

	got := strconv.FormatBool(c.value)
	c.ctx.CheckExpectation(
		c.value == v,
		func() {
			c.ctx.Fail("Expected values to be different\nGot: %s", got)
		},
		func() {
			c.ctx.Fail("Expected values to match\nGot: %s\nWant: %s", got, strconv.FormatBool(v))
		},
	)
}

func (c booleanChecker) ToBeFalsy() {
	c.ctx.t.Helper()

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

func (c booleanChecker) ToBeTruthy() {
	c.ctx.t.Helper()

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
