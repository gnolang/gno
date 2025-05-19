package expect

type ErrorChecker interface {
	Not() ErrorChecker
	WithMessage(string)
	WithError(error)
}

type errorChecker struct {
	ctx context
	err error
}

func (c errorChecker) Not() ErrorChecker {
	c.ctx.not = !c.ctx.not
	return c
}

func (c errorChecker) WithMessage(msg string) {
	c.ctx.t.Helper()

	if c.err == nil {
		c.ctx.Fail("Expected an error with message\nGot: nil\nWant: %s", msg)
		return
	}

	messageChecker{
		ctx:    c.ctx,
		msg:    c.err.Error(),
		prefix: "error",
	}.WithMessage(msg)
}

func (c errorChecker) WithError(err error) {
	c.ctx.t.Helper()

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
