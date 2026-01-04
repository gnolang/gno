package main

func main() {
	a := "1234"
	b := "1234"

	cond := a == b
	println(cond)
	println(cond == (a == b))
	println((a == b) == cond)
	println((a == b) == (a == b))
	println(cond && (a == b))
	println(cond || (a > b))

}

// Output:
// true
// true
// true
// true
// true
// true
