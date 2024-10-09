package main

func main() {
	// test against uninitialized value: https://github.com/gnolang/gno/pull/1132
	var x string
	y := "Hello"
	results := [...]bool{
		x < y,
	}
	println(results)
}

// Output:
// array[(true bool)]
