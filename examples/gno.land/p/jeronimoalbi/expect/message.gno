package expect

type MessageChecker interface {
	Not() MessageChecker
	WithMessage(string)
}

type messageChecker struct {
	ctx    context
	msg    string
	prefix string
}

func (c messageChecker) Not() MessageChecker {
	c.ctx.not = !c.ctx.not
	return c
}

func (c messageChecker) WithMessage(msg string) {
	c.ctx.t.Helper()

	c.ctx.CheckExpectation(
		c.msg == msg,
		func() {
			c.ctx.Fail("Expected %s message to be different\nGot: %s", c.prefix, c.msg)
		},
		func() {
			c.ctx.Fail("Expected %s message to match\nGot: %s\nWant: %s", c.prefix, c.msg, msg)
		},
	)
}
