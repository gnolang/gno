package expect

import "gno.land/p/nt/ufmt"

const (
	MessageTypeCrossPanic MessageType = "cross panic"
	MessageTypeError                  = "error"
	MessageTypePanic                  = "panic"
)

// MessageType defines a type for message checker errors.
type MessageType string

// NewMessageChecker creates a new checker for text messages.
func NewMessageChecker(ctx Context, msg string, t MessageType) MessageChecker {
	return MessageChecker{ctx, msg, t}
}

// MessageChecker asserts text messages.
type MessageChecker struct {
	ctx     Context
	msg     string
	msgType MessageType
}

// Not negates the next called expectation.
func (c MessageChecker) Not() MessageChecker {
	c.ctx.negated = !c.ctx.negated
	return c
}

// WithMessage asserts that a message is equal to an expected message.
func (c MessageChecker) WithMessage(msg string) {
	c.ctx.T().Helper()
	c.ctx.CheckExpectation(c.msg == msg, func(ctx Context) string {
		if !ctx.IsNegated() {
			return ufmt.Sprintf("Expected %s message to match\nGot: %s\nWant: %s", string(c.msgType), c.msg, msg)
		}
		return ufmt.Sprintf("Expected %s message to be different\nGot: %s", string(c.msgType), c.msg)
	})
}
