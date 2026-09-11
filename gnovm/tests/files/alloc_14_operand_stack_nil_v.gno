// MAXALLOC: 1000000
package main

// Regression: GC must count operand-stack slots whose V is nil.
//
// Each level of park parks 64 pending int operands (V == nil) on m.Values
// while recursing; at depth 128 that is ~8192 slots, ~327KB of TypedValue
// wrappers that runtime never charges. The second make pushes the total
// over MAXALLOC and GC runs. If GC recounted only non-nil V, the parked
// slots would vanish from the tally, the 700KB allocation would fit, and
// the program would print 1300000 with the Go heap well over the 1MB cap.
func park(n int) int {
	if n == 0 {
		return len(make([]byte, 600_000)) + len(make([]byte, 700_000))
	}
	return 0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (0 + (park(n - 1)))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))
}

func main() {
	println(park(128))
}

// Error:
// allocation limit exceeded
