package gnolang

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimOriginFile(t *testing.T) {
	t.Parallel()
	tt := []struct {
		input  string
		expect string
	}{
		{"/Users/x/gno/gnovm/pkg/gnolang/frame.go", "gnovm/pkg/gnolang/frame.go"},
		{"/home/runner/work/gno/gno/gnovm/cmd/gno/run.go", "gnovm/cmd/gno/run.go"},
		{"/usr/local/go/src/runtime/debug/stack.go", "src/runtime/debug/stack.go"},
		{"/no/match/here.go", "here.go"},
		{"justafile.go", "justafile.go"},
		{"", ""},
	}
	for _, tc := range tt {
		assert.Equal(t, tc.expect, TrimOriginFile(tc.input),
			"TrimOriginFile(%q)", tc.input)
	}
}

func makeExceptionFromHere() *Exception {
	return NewException(typedString("boom"))
}

func TestNewException_CapturesGoStack(t *testing.T) {
	t.Parallel()
	ex := makeExceptionFromHere()
	if ex.GoStack == "" {
		t.Fatal("expected non-empty GoStack")
	}
	// First recorded frame should be the caller of NewException —
	// makeExceptionFromHere — not NewException itself.
	if !strings.Contains(ex.GoStack, "makeExceptionFromHere") {
		t.Fatalf("GoStack missing raise-site frame:\n%s", ex.GoStack)
	}
	if strings.Contains(ex.GoStack, ".NewException\n") {
		t.Fatalf("GoStack should skip NewException itself:\n%s", ex.GoStack)
	}
	// Captured frames are runtime-filtered.
	for _, line := range strings.Split(ex.GoStack, "\n") {
		if strings.HasPrefix(line, "runtime.") {
			t.Fatalf("unexpected runtime.* frame:\n%s", ex.GoStack)
		}
	}
}

func TestException_Error_Abort(t *testing.T) {
	t.Parallel()
	ex := &Exception{
		Abort:      true,
		Descriptor: "joined-chain message",
		Value:      typedString("value-side"),
	}
	assert.Equal(t, "joined-chain message", ex.Error())
}

func TestException_Error_AbortWithoutDescriptor(t *testing.T) {
	t.Parallel()
	// Abort flagged but Descriptor not yet populated — falls back to
	// Value.String() (typed-value form, not the joined-chain shape).
	ex := &Exception{Abort: true, Value: typedString("value-side")}
	assert.Equal(t, `("value-side" string)`, ex.Error())
}

func TestException_Error_NonAbort(t *testing.T) {
	t.Parallel()
	// Non-abort Error() returns Value.String() — typed-value form.
	ex := &Exception{Value: typedString("raised but not terminal")}
	assert.Equal(t, `("raised but not terminal" string)`, ex.Error())
}
