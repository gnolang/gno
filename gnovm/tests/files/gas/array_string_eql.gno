// MAXALLOC: 500000000
package main

// Test arrays of strings where both per-element and
// per-byte gas charges interact.

func main() {
	buf := make([]byte, 500)
	s := string(buf)
	a1 := [10]string{s, s, s, s, s, s, s, s, s, s}
	a2 := [10]string{s, s, s, s, s, s, s, s, s, s}
	_ = a1 == a2
}

// Gas:
// 813371
