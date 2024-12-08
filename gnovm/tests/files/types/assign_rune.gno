package main

var r rune

// TODO: print right dec, float
func main() {
	r = 'a'
	println(r)
	println(float32(r))
	println(float64(r))
	println(string(r))
}

// Output:
// 97
// 97
// 97
// a
