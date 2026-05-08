package vm

import (
	"errors"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	cmnerrors "github.com/gnolang/gno/tm2/pkg/errors"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
)

func TestBoundedString_Nil(t *testing.T) {
	assert.Equal(t, "<nil>", boundedString(nil, 0))
}

func TestBoundedString_String_Small(t *testing.T) {
	assert.Equal(t, "hello", boundedString("hello", 0))
}

func TestBoundedString_String_Truncated(t *testing.T) {
	huge := strings.Repeat("a", maxBoundedBytes*2)
	got := boundedString(huge, 0)
	assert.Equal(t, maxBoundedBytes+3, len(got)) // "..." suffix
	assert.True(t, strings.HasSuffix(got, "..."))
}

func TestBoundedString_Bytes(t *testing.T) {
	huge := make([]byte, maxBoundedBytes*2)
	for i := range huge {
		huge[i] = 'x'
	}
	got := boundedString(huge, 0)
	assert.Equal(t, maxBoundedBytes+3, len(got))
}

func TestBoundedString_UnhandledPanicError_Short(t *testing.T) {
	up := gno.UnhandledPanicError{Descriptor: "panic message"}
	assert.Equal(t, "panic message", boundedString(up, 0))
}

func TestBoundedString_UnhandledPanicError_Truncated(t *testing.T) {
	up := gno.UnhandledPanicError{Descriptor: strings.Repeat("a", maxBoundedBytes*2)}
	got := boundedString(up, 0)
	assert.Equal(t, maxBoundedBytes+3, len(got))
}

func TestBoundedString_Exception(t *testing.T) {
	e := &gno.Exception{
		Value: gno.TypedValue{T: gno.StringType, V: gno.StringValue("oops")},
	}
	got := boundedString(e, 0)
	assert.Equal(t, "oops", got)
}

func TestBoundedString_OutOfGas(t *testing.T) {
	oog := stypes.OutOfGasError{Descriptor: "out of gas"}
	got := boundedString(oog, 0)
	assert.Contains(t, got, "out of gas")
}

func TestBoundedString_CmnError_FmtError(t *testing.T) {
	// errors.New produces a cmnError with FmtError data.
	err := cmnerrors.New("VM panic: %s", "details")
	got := boundedString(err, 0)
	// Should return the raw format string ("VM panic: %s") without
	// invoking Sprintf on args.
	assert.Equal(t, "VM panic: %s", got)
}

func TestBoundedString_CmnError_HugeFmtUnaffected(t *testing.T) {
	// If the format string itself is huge, it gets truncated via the
	// raw format extraction (no Sprintf on potentially-huge args).
	huge := strings.Repeat("a", maxBoundedBytes*2)
	err := cmnerrors.New("%s", huge)
	got := boundedString(err, 0)
	// Result is the format "%s" — short, no truncation.
	assert.Equal(t, "%s", got)
}

func TestBoundedString_CmnError_Wrapped(t *testing.T) {
	// Wrap an unhandled-panic-like error.
	inner := gno.UnhandledPanicError{Descriptor: "inner"}
	err := cmnerrors.Wrap(inner, "outer")
	got := boundedString(err, 0)
	// cmnError.Data() returns the wrapped error (UnhandledPanicError);
	// FmtError assertion fails; unwrap returns the wrapped error;
	// recurse → matches gno.UnhandledPanicError → returns Descriptor.
	assert.Equal(t, "inner", got)
}

func TestBoundedString_UnknownError(t *testing.T) {
	err := errors.New("stdlib error")
	got := boundedString(err, 0)
	// stdlib *errors.errorString implements error but doesn't have
	// Unwrap; doesn't implement cmnerrors.Error. Falls through to
	// generic error arm. errors.Unwrap returns nil → <error: %T>.
	assert.Contains(t, got, "<error:")
}

func TestBoundedString_DepthLimit(t *testing.T) {
	// Build a wrap chain that re-wraps repeatedly. cmnError.Wrap of
	// itself will cause deep recursion in boundedString via Unwrap.
	var err error = cmnerrors.New("base")
	for i := 0; i < 20; i++ {
		err = cmnerrors.Wrap(err, "layer")
	}
	got := boundedString(err, 0)
	// Eventually hits depth limit OR resolves to a FmtError. Either
	// is acceptable; just ensure no panic and bounded output.
	assert.LessOrEqual(t, len(got), maxBoundedBytes+3)
}

func TestBoundedString_Default(t *testing.T) {
	type myCustom struct{ x int }
	got := boundedString(myCustom{42}, 0)
	assert.Contains(t, got, "vm.myCustom")
}

func TestTruncate_Boundary(t *testing.T) {
	assert.Equal(t, "abc", truncate("abc"))
	assert.Equal(t, strings.Repeat("a", maxBoundedBytes), truncate(strings.Repeat("a", maxBoundedBytes)))
	assert.Equal(t,
		strings.Repeat("a", maxBoundedBytes)+"...",
		truncate(strings.Repeat("a", maxBoundedBytes+1)),
	)
}
