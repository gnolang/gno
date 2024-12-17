package main

func main() {
	printCap(nil)
}

func printCap(arr *[2]int) {
	println(cap(arr))
}

// Output:
// 2
