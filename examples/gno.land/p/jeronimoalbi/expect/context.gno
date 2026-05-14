package expect

import (
	"strings"

	"gno.land/p/nt/ufmt/v0"
)

const defaultAssertFailMsg = "assert expectation failed"

// NewContext creates a new testing context.
func NewContext(t TestingT) Context {
	return Context{t: t}
}

// Context preserves the current testing context.
type Context struct {
	t       TestingT
	negated bool
	prefix  string
}

// T returns context's testing T instance.
func (c Context) T() TestingT {
	if c.t == nil {
		panic("expect: context is not initialized")
	}
	return c.t
}

// Prefix returns context's error prefix.
func (c Context) Prefix() string {
	return c.prefix
}

// IsNegated checks if current context negates current assert expectations.
func (c Context) IsNegated() bool {
	return c.negated
}

// CheckExpectation checks an assert expectation and calls a callback on fail.
// It returns true when the asserted expectation fails.
// Callback is called when a negated assertion succeeds or when non negated assertion fails.
func (c Context) CheckExpectation(success bool, cb func(Context) string) bool {
	failed := (c.negated && success) || (!c.negated && !success)
	if failed {
		msg := cb(c)
		if strings.TrimSpace(msg) == "" {
			msg = defaultAssertFailMsg
		}

		c.Fail(msg)
	}
	return failed
}

// Fail makes the current test fail with a custom message.
func (c Context) Fail(msg string, args ...any) {
	if c.prefix != "" {
		msg = c.prefix + " - " + msg
	}

	c.t.Fatalf(msg, args...)
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
