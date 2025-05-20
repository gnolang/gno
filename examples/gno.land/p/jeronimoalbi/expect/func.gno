package expect

type (
	Fn         = func()
	ErrorFn    = func() error
	AnyFn      = func() any
	AnyErrorFn = func() (any, error)
)

func Func(t TestingT, fn any) FuncChecker {
	return FuncChecker{
		ctx: NewContext(t),
		fn:  fn,
	}
}

type FuncChecker struct {
	ctx Context
	fn  any
}

func (c FuncChecker) WithFailPrefix(prefix string) FuncChecker {
	c.ctx.prefix = prefix
	return c
}

func (c FuncChecker) Not() FuncChecker {
	c.ctx.not = !c.ctx.not
	return c
}

func (c FuncChecker) ToFail() ErrorChecker {
	c.ctx.t().Helper()

	var err error
	switch fn := c.fn.(type) {
	case ErrorFn:
		err = fn()
	case AnyErrorFn:
		_, err = fn()
	default:
		c.ctx.Fail("Unsupported error func type\nGot: %T", c.fn)
		return ErrorChecker{}
	}

	c.ctx.CheckExpectation(
		err != nil,
		func() {
			c.ctx.Fail("Func failed with error\nGot: %s", err.Error())
		},
		func() {
			c.ctx.Fail("Expected func to return an error")
		},
	)

	return NewErrorChecker(c.ctx, err)
}

func (c FuncChecker) ToPanic() MessageChecker {
	c.ctx.t().Helper()

	var (
		msg      string
		panicked bool
	)

	// TODO: Can't use a switch because it triggers the following VM error:
	// "panic: should not happen, should be heapItemType: fn<()~VPBlock(1,0)>"
	//
	// case Fn:
	// 	msg, panicked = handlePanic(fn)
	// case ErrorFn:
	// 	msg, panicked = handlePanic(func() { err = fn() })
	// case AnyFn:
	// 	msg, panicked = handlePanic(func() { v = fn() })
	// case AnyErrorFn:
	// 	msg, panicked = handlePanic(func() { v, err = fn() })
	// default:
	// 	c.ctx.Fail("Unsupported func type\nGot: %T", c.fn)
	// }

	if fn, ok := c.fn.(Fn); ok {
		msg, panicked = handlePanic(fn)
	} else if fn, ok := c.fn.(ErrorFn); ok {
		msg, panicked = handlePanic(func() { _ = fn() })
	} else if fn, ok := c.fn.(AnyFn); ok {
		msg, panicked = handlePanic(func() { _ = fn() })
	} else if fn, ok := c.fn.(AnyErrorFn); ok {
		msg, panicked = handlePanic(func() { _, _ = fn() })
	} else {
		c.ctx.Fail("Unsupported func type\nGot: %T", c.fn)
		return MessageChecker{}
	}

	c.ctx.CheckExpectation(
		panicked,
		func() {
			c.ctx.Fail("Expected func not to panic\nGot: %s", msg)
		},
		func() {
			c.ctx.Fail("Expected function to panic")
		},
	)

	return NewMessageChecker(c.ctx, msg, MessageTypePanic)
}

func (c FuncChecker) ToReturn(value any) {
	c.ctx.t().Helper()

	var (
		err error
		v   any
	)

	if fn, ok := c.fn.(AnyFn); ok {
		v = fn()
	} else if fn, ok := c.fn.(AnyErrorFn); ok {
		v, err = fn()
	} else {
		c.ctx.Fail("Unsupported func type\nGot: %T", c.fn)
		return
	}

	if err != nil {
		c.ctx.Fail("Function returned unexpected error\nGot: %s", err.Error())
		return
	}

	if c.ctx.not {
		Value(c.ctx.t(), v).Not().ToEqual(value)
	} else {
		Value(c.ctx.t(), v).ToEqual(value)
	}
}

func handlePanic(f func()) (msg string, panicked bool) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		panicked = true

		if err, ok := r.(error); ok {
			msg = err.Error()
			return
		}

		if s, ok := r.(string); ok {
			msg = s
			return
		}

		msg = "unsupported panic type"
	}()

	f()
	return
}
