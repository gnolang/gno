// Verifies the `// Verdict: Unsupported <tag>: <reason>` flow: the file
// is skipped before any execution. The body intentionally uses a feature
// Gno doesn't support (channels); without the line, verify refuses the
// file for lacking a verdict, and a Clean or Known* verdict would fail as
// stale, since verify never skips on Gno's error alone.

package main

func main() {
	ch := make(chan int, 1)
	ch <- 42
	println(<-ch)
}

// Verdict: Unsupported channel: channels not supported in Gno
