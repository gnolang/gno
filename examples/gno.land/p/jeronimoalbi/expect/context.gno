package expect

import (
	"strings"

	"gno.land/p/demo/ufmt"
)

func NewContext(t TestingT) Context {
	return Context{tt: t}
}

type Context struct {
	tt     TestingT
	not    bool
	prefix string
}

func (c Context) CheckExpectation(ok bool, fn, fn2 func()) {
	// Check must be false to pass when "not" is specified
	if ok && c.not {
		fn()
	}

	// Check must be true to pass when "not" is not specified
	if !ok && !c.not {
		fn2()
	}
}

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

type TestingT interface {
	Helper()
	Fatal(args ...any)
	Fatalf(format string, args ...any)
}

func MockTestingT(output *strings.Builder) TestingT {
	return &testingT{output}
}

type testingT struct{ buf *strings.Builder }

func (testingT) Helper()                          {}
func (t testingT) Fatal(args ...any)              { t.buf.WriteString(ufmt.Sprintln(args...)) }
func (t testingT) Fatalf(fmt string, args ...any) { t.buf.WriteString(ufmt.Sprintf(fmt+"\n", args...)) }
