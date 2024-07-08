package main

func pr(x bool) {
	println(x)
}

func main() {
	a := 1
	b := 1
	pr(a == b)
}

// Output:
// true
