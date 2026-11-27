package main

const limit = 10

// A local defined from the package const it shadows reads that const.
func main() {
	limit := limit * 2
	println(limit)
}

// Output:
// 20
