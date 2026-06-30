package main

func main() {
	for k, v := range map[any]int{[]int{1, 2}: 10} {
		println(k, v)
	}
}

// Stacktrace:
// panic: runtime error: hash of unhashable type []int
// main<VPBlock(1,0)>()
//     main/range13.gno:5

// Error:
// runtime error: hash of unhashable type []int
