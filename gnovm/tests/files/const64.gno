package main

func main() {
	const a = len([1]int{len([]int{1})})
	println(a)
}

// Error:
// main/files/const64.gno:4:8: [](const-type int){(const (1 int))} (variable of type []int) is not constant
