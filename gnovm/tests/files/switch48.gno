package main

type T struct{ n int }

// Clause names are re-created on fallthrough, even when the next clause
// declares the same number of names with different types.
func main() {
	switch 1 {
	case 1:
		a, b, c := &T{1}, "s", 3.5
		println(a.n, b, c)
		fallthrough
	case 2:
		var z *T
		println(z == nil)
	}

	switch 1 {
	case 1:
		x := "one"
		_ = x
		fallthrough
	case 2:
		var y int
		println(y)
	}
}

// Output:
// 1 s 3.5
// true
// 0
