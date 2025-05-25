package expect

import (
	"strings"

	"gno.land/p/demo/ufmt"
)

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

// CheckExpectation checks an assert expectation and calls one callback on fail.
// First callback is called when a non negated expectation fails.
// Otherwise second callback is called when a negated one fails.
func (c Context) CheckExpectation(success bool, fn, fnNot func()) {
	if !success && !c.negated {
		//  Expectation failed
		fn()
	} else if success && c.negated {
		// Negated expectation failed
		fnNot()
	}
}

// Fail makes the current test fail with a custom message.
func (c Context) Fail(msg string, args ...any) {
	if c.prefix != "" {
		msg = c.prefix + " - " + msg
	}

	c.tt.Fatalf(msg, args...)
}

func (c Context) t() TestingT {
	if c.tt == nil {
		panic("expect: context is not initialized")
	}
	return c.tt
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
