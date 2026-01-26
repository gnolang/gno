// MAXALLOC: 500000000
package main

type S struct {
	a, b, c int
}

func main() {
	var x [100]S
	_ = x
}

// Gas:
// 88541
