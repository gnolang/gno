// MAXALLOC: 500000000

package main

// Test that infinite recursion triggers "allocation limit exceeded"
// and NOT "should not happen, allocation limit exceeded while gc."
// This verifies that GC recount is consistent with allocator tracking.

func f() int {
	return f()
}

func main() {
	f()
}

// Error:
// allocation limit exceeded
