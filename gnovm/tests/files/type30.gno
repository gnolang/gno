package main

type String string

func equal(a, b string) bool {
	return a == b
}

func main() {
	x := "dontcare"
	if !equal(x, x) {
		panic("should not happen")
	}
	y := String(x)
	println(equal(x, y))
}

// Error:
// main/type30.gno:15:10-21: cannot use main.String as string without explicit conversion

// TypeCheckError:
// main/type30.gno:15:19: cannot use y (variable of string type String) as string value in argument to equal
