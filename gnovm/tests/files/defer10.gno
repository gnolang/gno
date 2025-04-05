package main

func Swap(values [3]int) {
	values[0] = 1337
	println(values)
}

func main() {
	values := [3]int{1, 2, 3}
	defer func() {
		println(values)
	}()
	defer Swap(values)

	println("test")
}

// Output:
// test
// array[(1337 int),(2 int),(3 int)]
// array[(1 int),(2 int),(3 int)]
