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
// The tail below is harness-written: `// GnoOutput:` pins Gno's
// stdout, `// GoOutput:` pins `go run`'s, and the `// Verdict:` line
// records whether they match. An intentional Gno-vs-Go divergence is
// blessed by a `// Verdict: KnownDivergence run: <reason>` line.

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

// Verdict: Clean run
