// MAXALLOC: 500000000
package main

// [1]int == [1]int
// large_array_eql.gno: same test with [100]int
func main() {
	a1 := [1]int{}
	a2 := [1]int{}
	_ = a1 == a2
}

// Gas:
// 5601
