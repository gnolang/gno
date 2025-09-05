package expect

import "gno.land/p/nt/ufmt"

// NewErrorChecker creates a new checker of errors.
func NewErrorChecker(ctx Context, err error) ErrorChecker {
	return ErrorChecker{ctx, err}
}

// ErrorChecker asserts error values.
type ErrorChecker struct {
	ctx Context
	err error
}

// Not negates the next called expectation.
func (c ErrorChecker) Not() ErrorChecker {
	c.ctx.negated = !c.ctx.negated
	return c
}

// WithMessage asserts that current error contains an expected message.
func (c ErrorChecker) WithMessage(msg string) {
	c.ctx.T().Helper()

	if c.err == nil {
		c.ctx.Fail("Expected an error with message\nGot: nil\nWant: %s", msg)
		return
	}

	NewMessageChecker(c.ctx, c.err.Error(), MessageTypeError).WithMessage(msg)
}

// WithError asserts that current error message is the same as an expected error.
func (c ErrorChecker) WithError(err error) {
	c.ctx.T().Helper()

	if c.err == nil {
		if err != nil {
			c.ctx.Fail("Expected an error\nGot: nil\nWant: %s", err.Error())
		}
		return
	}

	got := c.err.Error()
	c.ctx.CheckExpectation(got == err.Error(), func(ctx Context) string {
		if !ctx.IsNegated() {
			return ufmt.Sprintf("Expected errors to match\nGot: %s\nWant: %s", got, err.Error())
		}
		return ufmt.Sprintf("Expected errors to be different\nGot: %s", got)
	})
}
