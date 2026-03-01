package main

func main() {
	var a uint = 3

	b := 1<<a == uint(2)<<a

	println(b)
}
// Output:
// false