package main

// A zero untyped rune shifted by a constant count folds to 0, as before,
// for any count (gc rejects counts this large).

const C = '\x00' << 10001

func main() {
	println(C, '\x00'<<40, 'a'>>10001)
}

// Output:
// 0 0 0

// TypeCheckError:
// main/rune19.gno:6:21: invalid operation: invalid shift count 10001 (untyped int constant); main/rune19.gno:9:30: invalid operation: invalid shift count 10001 (untyped int constant)
