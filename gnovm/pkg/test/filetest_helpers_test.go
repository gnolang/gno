package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHarnessAfterRun_SlicesPastMachineRun(t *testing.T) {
	stack := strings.Join([]string{
		"goroutine 1 [running]:",
		"runtime/debug.Stack()",
		"\t/usr/local/go/src/runtime/debug/stack.go:24 +0x65",
		"github.com/gnolang/gno/gnovm/pkg/gnolang.(*Machine).Run(0xc000)",
		"\t/Users/x/gno/gnovm/pkg/gnolang/machine.go:1550 +0x65",
		"github.com/gnolang/gno/gnovm/pkg/test.(*TestOptions).runTest(...)",
		"\t/Users/x/gno/gnovm/pkg/test/filetest.go:497 +0x123",
		"",
	}, "\n")
	got := harnessAfterRun([]byte(stack))
	if strings.Contains(got, "(*Machine).Run(") {
		t.Fatalf("Machine.Run frame should be sliced out:\n%s", got)
	}
	if !strings.Contains(got, "runTest") {
		t.Fatalf("expected harness caller after slice:\n%s", got)
	}
	// Project-relative path conversion.
	if !strings.Contains(got, "\tgnovm/pkg/test/filetest.go:497") {
		t.Fatalf("expected trimmed path:\n%s", got)
	}
}

func TestHarnessAfterRun_NoMarker_FallsThrough(t *testing.T) {
	stack := "goroutine 1 [running]:\nsomething.Else()\n\t/Users/x/gno/gnovm/foo.go:1 +0x0\n"
	got := harnessAfterRun([]byte(stack))
	// No (*Machine).Run( so the whole stack is returned, paths trimmed.
	if !strings.Contains(got, "gnovm/foo.go:1") {
		t.Fatalf("expected trimmed path in passthrough:\n%s", got)
	}
}

func TestHarnessAfterRun_Empty(t *testing.T) {
	assert.Equal(t, "", harnessAfterRun(nil))
	assert.Equal(t, "", harnessAfterRun([]byte{}))
}

func TestTrimStackPaths(t *testing.T) {
	in := strings.Join([]string{
		"foo.Bar()",
		"\t/Users/x/gno/gnovm/pkg/gnolang/frame.go:42 +0x12",
		"baz.Qux()",
		"\t/usr/local/go/src/runtime/proc.go:1 +0x0",
		"",
	}, "\n")
	out := trimStackPaths(in)
	if !strings.Contains(out, "\tgnovm/pkg/gnolang/frame.go:42") {
		t.Fatalf("expected gnovm path trim:\n%s", out)
	}
	if !strings.Contains(out, "\tsrc/runtime/proc.go:1") {
		t.Fatalf("expected stdlib /src/ trim:\n%s", out)
	}
}

func TestGoOriginOrStack_PrefersVMChain(t *testing.T) {
	rr := runResult{
		GoVMStack:    "frame.Raise\n\tgnovm/pkg/gnolang/frame.go:1\n",
		GoPanicStack: []byte("...\n(*Machine).Run(...)\n\t/Users/x/gno/gnovm/pkg/gnolang/machine.go:1\nharness.Caller(...)\n\t/Users/x/gno/gnovm/pkg/test/filetest.go:1\n"),
	}
	got := goOriginOrStack(rr)
	if !strings.HasPrefix(got, "\ngo stack:\nframe.Raise") {
		t.Fatalf("expected go stack to lead with VM chain:\n%s", got)
	}
	if !strings.Contains(got, "harness.Caller") {
		t.Fatalf("expected harness suffix after VM chain:\n%s", got)
	}
}

func TestGoOriginOrStack_FallsBackToRawDump(t *testing.T) {
	rr := runResult{
		GoPanicStack: []byte("some raw go stack"),
	}
	got := goOriginOrStack(rr)
	assert.Equal(t, "\nstack:\nsome raw go stack", got)
}

func TestGoOriginOrStack_EmptyWhenNothing(t *testing.T) {
	assert.Equal(t, "", goOriginOrStack(runResult{}))
}
