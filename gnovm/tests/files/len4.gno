package main

func main() {
	printLen(nil)
}

func printLen(arr *[2]int) {
	println(len(arr))
}

// Output:
// 2
