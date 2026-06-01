// Verifies the filetest walker accepts plain .go files under
// gnovm/tests/files/gocorpus/testdata/. Letting Go-source files
// drop in without rename means a per-fix regression test for any
// file from Go's standard test corpus is just `cp` + commit.
//
// The .go extension is invisible to `go list` / `go build` here
// because Go's tooling ignores any directory named `testdata/`;
// the walker additionally enforces that `.go` filetests must live
// under such a segment.
//
// This file ships with no `// Output:` directive — the harness
// auto-derives the expected output by running it through `go run`
// and compares Gno's output to that. To bless an intentional
// Gno-vs-Go divergence, add an explicit `// Output:` line with
// Gno's actual output; the directive bypasses the auto-derive
// step and serves as documentation.

package main

import "fmt"

func main() {
	// fmt.Println goes to stdout in both Go and Gno; the builtin
	// `println` would go to stderr in Go but stdout in Gno — a real
	// divergence we don't want the canary to surface.
	fmt.Println("canary: .go filetest accepted")
}

// GnoOutput:
// canary: .go filetest accepted

// GoOutput:
// canary: .go filetest accepted
