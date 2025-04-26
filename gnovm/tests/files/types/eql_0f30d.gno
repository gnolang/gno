package main

func main() {
	println([]int{1} == []int{1})
}

// Error:
// main/files/types/eql_0f30d.gno:4:10: []int can only be compared to nil

// TypeCheckError:
// main/files/types/eql_0f30d.gno:4:10: invalid operation: []int{…} == []int{…} (slice can only be compared to nil)
