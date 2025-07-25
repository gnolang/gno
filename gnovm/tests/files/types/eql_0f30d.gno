package main

func main() {
	println([]int{1} == []int{1})
}

// Error:
// main/eql_0f30d.gno:4:10-30: []int can only be compared to nil

// TypeCheckError:
// main/eql_0f30d.gno:4:10: invalid operation: []int{…} == []int{…} (slice can only be compared to nil)
