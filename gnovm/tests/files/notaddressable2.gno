package main

func main() {
	arr := &getArray()[0]
	println(arr)
}

func getArray() [3]int {
	return [3]int{1, 2, 3}
}

// Error:
// doOpRef: expr not addressable: array[(1 int),(2 int),(3 int)]