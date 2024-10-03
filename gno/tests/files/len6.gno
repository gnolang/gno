package main

func main() {
	printLenCap(nil)
}

func printLenCap(arr *[4]float64) {
	println(len(arr))
	println(cap(arr))
}

// Output:
// 4
// 4
