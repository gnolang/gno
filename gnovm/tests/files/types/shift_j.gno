package main

func main() {
	var a uint = 1

	b := 1<<a == 2<<a

	println(b)
}
// Output:
// false