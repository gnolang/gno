package expect

type (
	// Fn defines a type for generic functions.
	Fn = func()

	// ErrorFn defines a type for generic functions that return an error.
	ErrorFn = func() error

	// AnyFn defines a type for generic functions that returns a value.
	AnyFn = func() any

	// AnyErrorFn defines a type for generic functions that return a value and an error.
	AnyErrorFn = func() (any, error)
)

// Func creates a new checker for functions.
func Func(t TestingT, fn any) FuncChecker {
	return FuncChecker{
		ctx: NewContext(t),
		fn:  fn,
	}
}

// FuncChecker asserts function panics, errors and returned value.
type FuncChecker struct {
	ctx Context
	fn  any
}

// WithFailPrefix assigns a prefix that will be prefixed to testing errors when an assertion fails.
func (c FuncChecker) WithFailPrefix(prefix string) FuncChecker {
	c.ctx.prefix = prefix
	return c
}

// Not negates the next called expectation.
func (c FuncChecker) Not() FuncChecker {
	c.ctx.negated = !c.ctx.negated
	return c
}

// ToFail return an error checker to assert if current function returns an error.
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
			c.ctx.Fail("Expected func to return an error")
		},
		func() {
			c.ctx.Fail("Func failed with error\nGot: %s", err.Error())
		},
	)

	return NewErrorChecker(c.ctx, err)
}

// ToPanic return an message checker to assert if current function panicked.
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
			c.ctx.Fail("Expected function to panic")
		},
		func() {
			c.ctx.Fail("Expected func not to panic\nGot: %s", msg)
		},
	)

	return NewMessageChecker(c.ctx, msg, MessageTypePanic)
}

// ToReturn asserts that current function returned a value equal to an expected value.
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

	if c.ctx.negated {
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
