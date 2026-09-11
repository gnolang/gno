// MAXALLOC: 1000000
package main

// Regression: GC must visit the operand stack (m.Values).
//
// Each big() result is pushed onto the operand stack and stays there until
// doOpSliceLit pops all elements to build the slice. When the 4th element
// pushes total allocation over MAXALLOC, GC runs while the first three
// (~900KB) are reachable *only* from the operand stack. If GC skipped
// m.Values, Reset()+Recount() would drop them from the tally, the 4th
// allocation would succeed against phantom headroom, and the program would
// print 4 — with the Go heap holding well over the 1MB cap.
func big() []byte { return make([]byte, 300_000) }

func main() {
	s := [][]byte{big(), big(), big(), big()}
	println(len(s))
}

// Error:
// allocation limit exceeded
