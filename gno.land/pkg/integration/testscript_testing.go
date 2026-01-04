package integration

import (
	"errors"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This error is from testscript.Fatalf and is needed to correctly
// handle the FailNow method.
// see: https://github.com/rogpeppe/go-internal/blob/32ae33786eccde1672d4ba373c80e1bc282bfbf6/testscript/testscript.go#L799-L812
var errFailNow = errors.New("fail now")

var (
	_ require.TestingT = (*testingTS)(nil)
	_ assert.TestingT  = (*testingTS)(nil)
	_ TestingTS        = &testing.T{}
)

type TestingTS = require.TestingT

type testingTS struct {
	*testscript.TestScript
}

func TSTestingT(ts *testscript.TestScript) TestingTS {
	return &testingTS{ts}
}

func (t *testingTS) Errorf(format string, args ...any) {
	defer recover() // we can ignore recover result, we just want to catch it up
	t.Fatalf(format, args...)
}

func (t *testingTS) FailNow() {
	// unfortunately we can't access underlying `t.t.FailNow` method
	panic(errFailNow)
}
