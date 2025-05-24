package expect

func Value(t TestingT, value any) ValueChecker {
	return ValueChecker{
		ctx:   NewContext(t),
		value: value,
	}
}

type ValueChecker struct {
	ctx   Context
	value any
}

func (c ValueChecker) WithFailPrefix(prefix string) ValueChecker {
	c.ctx.prefix = prefix
	return c
}

func (c ValueChecker) Not() ValueChecker {
	c.ctx.not = !c.ctx.not
	return c
}

func (c ValueChecker) ToBeNil() {
	c.ctx.t().Helper()
	c.ctx.CheckExpectation(
		c.value == nil,
		func() {
			c.ctx.Fail("Expected a non nil value")
		},
		func() {
			c.ctx.Fail("Expected value to be nil\nGot: %v", c.value)
		},
	)
}

func (c ValueChecker) ToEqual(value any) {
	c.ctx.t().Helper()

	switch v := value.(type) {
	case string:
		c.AsString().ToEqual(v)
	case []byte:
		c.AsString().ToEqual(string(v))
	case bool:
		c.AsBoolean().ToEqual(v)
	case float32:
		c.AsFloat().ToEqual(float64(v))
	case float64:
		c.AsFloat().ToEqual(v)
	case uint:
		c.AsUint().ToEqual(uint64(v))
	case uint8:
		c.AsUint().ToEqual(uint64(v))
	case uint16:
		c.AsUint().ToEqual(uint64(v))
	case uint32:
		c.AsUint().ToEqual(uint64(v))
	case uint64:
		c.AsUint().ToEqual(v)
	case int:
		c.AsInt().ToEqual(int64(v))
	case int8:
		c.AsInt().ToEqual(int64(v))
	case int16:
		c.AsInt().ToEqual(int64(v))
	case int32:
		c.AsInt().ToEqual(int64(v))
	case int64:
		c.AsInt().ToEqual(v)
	default:
		c.ctx.Fail("Unsupported type: %T", value)
	}
}

func (c ValueChecker) AsString() StringChecker {
	c.ctx.t().Helper()

	v, err := asString(c.value)
	if err != nil {
		c.ctx.Fail("Failed: %s: expected a string value\nGot: %T", err.Error(), c.value)
		return StringChecker{}
	}

	return NewStringChecker(c.ctx, v)
}

func (c ValueChecker) AsBoolean() BooleanChecker {
	c.ctx.t().Helper()

	v, err := asBoolean(c.value)
	if err != nil {
		c.ctx.Fail("Failed: %s: expected a boolean value\nGot: %T", err.Error(), c.value)
		return BooleanChecker{}
	}

	return NewBooleanChecker(c.ctx, v)
}

func (c ValueChecker) AsFloat() FloatChecker {
	c.ctx.t().Helper()

	v, err := asFloat(c.value)
	if err != nil {
		c.ctx.Fail("%s: expected a float value\nGot: %T", err.Error(), c.value)
		return FloatChecker{}
	}

	return NewFloatChecker(c.ctx, v)
}

func (c ValueChecker) AsUint() UintChecker {
	c.ctx.t().Helper()

	v, err := asUint(c.value)
	if err != nil {
		c.ctx.Fail("Failed: %s: expected a uint value\nGot: %T", err.Error(), c.value)
		return UintChecker{}
	}

	return NewUintChecker(c.ctx, v)
}

func (c ValueChecker) AsInt() IntChecker {
	c.ctx.t().Helper()

	v, err := asInt(c.value)
	if err != nil {
		c.ctx.Fail("Failed: %s: expected an int value\nGot: %T", err.Error(), c.value)
		return IntChecker{}
	}

	return NewIntChecker(c.ctx, v)
}
