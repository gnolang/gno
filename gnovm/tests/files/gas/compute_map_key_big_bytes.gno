// MAXALLOC: 50000000
package main

func main() {
	m := make(map[[1<<25]byte]int)
	// This triggers a ComputeMapKey call on a large array
	// We should have different gas depending on the size of the array
	m[[1<<25]byte{'0'}] = 1
}


// Gas:
// 134224443
