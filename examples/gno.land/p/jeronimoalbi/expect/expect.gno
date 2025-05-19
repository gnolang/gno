package expect

import (
	"errors"
	"strings"

	"gno.land/p/demo/ufmt"
)

var ErrIncompatibleType = errors.New("incompatible type")

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
