package expect

import (
	"strings"

	"gno.land/p/demo/ufmt"
)

const defaultAssertFailMsg = "assert expectation failed"

// NewContext creates a new testing context.
func NewContext(t TestingT) Context {
	return Context{tt: t}
}

// Context preserves the current testing context.
type Context struct {
	tt      TestingT
	negated bool
	prefix  string
}

// T returns context's testing T instance.
func (c Context) T() TestingT {
	if c.tt == nil {
		panic("expect: context is not initialized")
	}
	return c.tt
}

// IsNegated checks if current context negates current assert expectations.
func (c Context) IsNegated() bool {
	return c.negated
}

// CheckExpectation checks an assert expectation and calls a callback on fail.
// Callback is called when a negated assertion succeeds or when non negated assertion fails.
func (c Context) CheckExpectation(success bool, cb func(Context) string) {
	if failed := (c.negated && success) || (!c.negated && !success); failed {
		msg := cb(c)
		if strings.TrimSpace(msg) == "" {
			msg = defaultAssertFailMsg
		}

		c.Fail(msg)
	}
}

// Fail makes the current test fail with a custom message.
func (c Context) Fail(msg string, args ...any) {
	if c.prefix != "" {
		msg = c.prefix + " - " + msg
	}

	c.tt.Fatalf(msg, args...)
}

// TestingT defines a minimal interface for `testing.T` instances.
type TestingT interface {
	Helper()
	Fatal(args ...any)
	Fatalf(format string, args ...any)
}

// MockTestingT creates a new testing mock that writes testing output to a string builder.
func MockTestingT(output *strings.Builder) TestingT {
	return &testingT{output}
}

type testingT struct{ buf *strings.Builder }

func (testingT) Helper()                          {}
func (t testingT) Fatal(args ...any)              { t.buf.WriteString(ufmt.Sprintln(args...)) }
func (t testingT) Fatalf(fmt string, args ...any) { t.buf.WriteString(ufmt.Sprintf(fmt+"\n", args...)) }
