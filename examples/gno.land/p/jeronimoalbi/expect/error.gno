package expect

func NewErrorChecker(ctx Context, err error) ErrorChecker {
	return ErrorChecker{ctx, err}
}

type ErrorChecker struct {
	ctx Context
	err error
}

func (c ErrorChecker) Not() ErrorChecker {
	c.ctx.not = !c.ctx.not
	return c
}

func (c ErrorChecker) WithMessage(msg string) {
	c.ctx.t().Helper()

	if c.err == nil {
		c.ctx.Fail("Expected an error with message\nGot: nil\nWant: %s", msg)
		return
	}

	NewMessageChecker(c.ctx, c.err.Error(), MessageTypeError).WithMessage(msg)
}

func (c ErrorChecker) WithError(err error) {
	c.ctx.t().Helper()

	if c.err == nil {
		if err != nil {
			c.ctx.Fail("Expected an error\nGot: nil\nWant: %s", err.Error())
		}
		return
	}

	got := c.err.Error()
	c.ctx.CheckExpectation(
		got == err.Error(),
		func() {
			c.ctx.Fail("Expected errors to be different\nGot: %s", got)
		},
		func() {
			c.ctx.Fail("Expected errors to match\nGot: %s\nWant: %s", got, err.Error())
		},
	)
}
