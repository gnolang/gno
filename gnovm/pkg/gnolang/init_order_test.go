package gnolang_test

import (
	"fmt"
	"io"
	"path/filepath"
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
