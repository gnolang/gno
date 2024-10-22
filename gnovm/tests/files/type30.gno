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
// main/files/type30.gno:15:10: cannot use main.String as string without explicit conversion
