// Verifies the `// Verdict: KnownIssue run: #NNNN <reason>` flow: Gno is
// wrong, Go is right, the file pins both sides and cites the tracking
// issue. Verify fails the moment the bug is fixed ("stale verdict"), so
// the fix PR flips this line to `Clean run` and closes the issue.
//
// The bug: a method expression on an interface type, I.M, is rejected
// at preprocess; Go accepts it and calls the dynamic method.

package main

type I interface{ M() int }

type T int

func (T) M() int { return 1 }

func main() {
	f := I.M
	println(f(T(0)))
}

// GnoOutput:

// GnoError:
// main/knownissue_canary.go:18:7-10: unknown *DeclaredType method named M

// GoOutput:
// 1

// Verdict: KnownIssue run: #5787 method expressions on interface types are rejected at preprocess
