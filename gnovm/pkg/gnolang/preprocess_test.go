package gnolang_test

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/stretchr/testify/require"
)

// TestInitOrderDeterminism verifies that package-level variable initialization
// always produces the same result across many runs, regardless of Go's
// non-deterministic map iteration over internal dependency sets.
// It runs a program with complex cross-variable dependencies 100 times and
// checks that the initialization order is always Go-spec-compliant.
func TestInitOrderDeterminism(t *testing.T) {
	// This source has vars that all depend (transitively, through emit) on
	// 'events', but Z also depends on A,B,C,D,E. The Go spec mandates that
	// the earliest-in-source-order ready variable is initialized next.
	// Expected order: events, B, A, C, G, D, E, Z (emit("L")), F.
	const src = `package main

var events []string
func emit(s string) string { events = append(events, s); return s }
var (
	Z = A + "-" + B + "-" + C + "-" + emit("L") + D + "-" + E
	B = emit("B")
	A = emit("A")
	C = emit("C")
	G = emit("G")
	D = emit("D")
	E = emit("E")
	F = emit("F")
)
func main() {
	for _, e := range events { println(e) }
}

// Output:
// B
// A
// C
// G
// D
// E
// L
// F
`
	rootDir, err := filepath.Abs("../../../")
	require.NoError(t, err)

	newOpts := func() *test.TestOptions {
		opts := test.NewTestOptions(rootDir, io.Discard, io.Discard, nil)
		return opts
	}
	sharedOpts := newOpts()

	const iters = 100
	for i := 0; i < iters; i++ {
		t.Run(fmt.Sprintf("iter%d", i), func(t *testing.T) {
			_, _, err := sharedOpts.RunFiletest("init_order_det.gno", []byte(src), sharedOpts.TestStore)
			require.NoError(t, err, "init order non-determinism or mismatch detected on iteration %d", i)
		})
	}
}

// TestCircDepDeterminism verifies that circular-dependency error messages are
// produced in a deterministic order even when a declaration has multiple
// function dependencies (each of which independently reads the same variable).
//
// Without sorting: `var a = B() + C()` causes a's ATTR_DECL_DEPS to hold
// {B, C}. Because Go map iteration is non-deterministic, the DFS in
// findUnresolvedDeps could traverse B first ("circular dependency: a -> B")
// or C first ("circular dependency: a -> C"), making the panic message flaky.
//
// After the fix (sorting ATTR_DECL_DEPS keys before DFS), the message is
// always deterministic ("a -> B" because "B" < "C" lexicographically).
func TestCircDepDeterminism(t *testing.T) {
	const src = `package main

var a = B() + C()

func B() int { return a }

func C() int { return a }

func main() {}`

	rootDir, err := filepath.Abs("../../../")
	require.NoError(t, err)

	opts := test.NewTestOptions(rootDir, io.Discard, io.Discard, nil)

	const iters = 100
	var seen []string
	for i := 0; i < iters; i++ {
		_, _, runErr := opts.RunFiletest("circ_dep_det.gno", []byte(src), opts.TestStore)
		// A circular dep error is expected; extract just the "circular dependency: ..." line.
		if runErr == nil {
			t.Fatalf("iteration %d: expected circular dep error, got none", i)
		}
		errMsg := runErr.Error()
		idx := strings.Index(errMsg, "circular dependency: ")
		if idx < 0 {
			// May be a TypeCheckError mismatch with no circular dep in the message; skip.
			continue
		}
		end := strings.IndexByte(errMsg[idx:], '\n')
		var circdep string
		if end >= 0 {
			circdep = errMsg[idx : idx+end]
		} else {
			circdep = errMsg[idx:]
		}
		found := false
		for _, s := range seen {
			if s == circdep {
				found = true
				break
			}
		}
		if !found {
			seen = append(seen, circdep)
		}
	}
	if len(seen) > 1 {
		t.Errorf("circular dependency error message is non-deterministic across %d runs: %v", iters, seen)
	}
}
