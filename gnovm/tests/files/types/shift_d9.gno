package main

func main() {
	r := bool(1<<2+1 == 1>>2)
	println(r)
}

// Output:
// false
