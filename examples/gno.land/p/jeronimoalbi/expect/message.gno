package expect

const (
	MessageTypeError MessageType = "error"
	MessageTypePanic             = "panic"
)

type MessageType string

func NewMessageChecker(ctx Context, msg string, t MessageType) MessageChecker {
	return MessageChecker{ctx, msg, t}
}

type MessageChecker struct {
	ctx     Context
	msg     string
	msgType MessageType
}

func (c MessageChecker) Not() MessageChecker {
	c.ctx.not = !c.ctx.not
	return c
}

func (c MessageChecker) WithMessage(msg string) {
	c.ctx.t().Helper()
	c.ctx.CheckExpectation(
		c.msg == msg,
		func() {
			c.ctx.Fail("Expected %s message to be different\nGot: %s", string(c.msgType), c.msg)
		},
		func() {
			c.ctx.Fail("Expected %s message to match\nGot: %s\nWant: %s", string(c.msgType), c.msg, msg)
		},
	)
}
