package main

func main() {
	var a uint = 1

	b := (1<<a == 2<<a) == (1.0 == 2)  

	println(b)
}
// Output:
// true