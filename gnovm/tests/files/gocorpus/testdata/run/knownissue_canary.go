// run

// Verifies the `// Verdict: KnownIssue run: #NNNN <reason>` flow: Gno is
// wrong, Go is right, the file pins both sides and cites the tracking
// issue. Verify fails the moment the bug is fixed ("stale verdict"), so
// the fix PR flips this line to `Clean run` and closes the issue.
//
// The bug: the builtin print inserts spaces between its operands; Go
// writes them back to back (only println adds spaces).

package main

func main() {
	print("ab", "cd", "\n")
}

// GnoOutput:
// ab cd

// GoOutput:
// abcd

// Verdict: KnownIssue run: #6331 builtin print inserts spaces between operands
