// MAXALLOC: 5000000
package main

func main() {
	m := make(map[[1<<5]byte]int)
	// This triggers a ComputeMapKey call on a large array
	// We should have different gas depending on the size of the array
	m[[1<<5]byte{'0'}] = 1
}


// Gas:
// 6687
