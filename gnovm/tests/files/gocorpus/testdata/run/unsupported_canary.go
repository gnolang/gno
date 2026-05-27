// Verifies the `// Unsupported:` directive: the file declares a
// reason at the bottom (same placement convention as `// Output:`)
// and is skipped before any execution. The body intentionally uses
// a feature Gno doesn't support (channels) — if the directive
// failed to take effect, the test would fail loudly.

package main

func main() {
	ch := make(chan int, 1)
	ch <- 42
	println(<-ch)
}

// Unsupported: channels not supported in Gno
