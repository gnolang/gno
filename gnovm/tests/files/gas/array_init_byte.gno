// MAXALLOC: 500000000
package main

func main() {
	var x [10000]byte
	_ = x
}

// Gas:
// 33962
