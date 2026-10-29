package main

// Comparisons: both operands take the common kind's default type.
func main() {
	var s uint = 32
	println(('a' << s) < (1 << s))
	println(('a' << s) < 100)
	println(('a' << s) == (1 << s))
	println((1 << s) == 0)
	println((1 << s) < (1 << 40))
	var i64 int64 = 1
	println(('a' << s) > i64)
	switch ('a' << s) < 100 {
	case true:
		println("t")
	default:
		println("f")
	}
	switch 1 + ('a' << s) {
	case 1:
		println("rune 1")
	default:
		println("other")
	}
	s = 0
	switch 1 + ('a' << s) {
	case 'b':
		println("rune b")
	}
	if ('a' << s) < 100 {
		println("lt")
	}
	b := ('a' << s) < (1 << s)
	println(b)
}

// Output:
// false
// true
// true
// false
// true
// true
// t
// rune 1
// rune b
// lt
// false
