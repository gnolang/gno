package main

// Comparisons of untyped rune constants use the exact values.

const c = '\U0010ffff' * '\U0010ffff'

func main() {
	println(c > 0, c == '\U0010ffff'*'\U0010ffff', c != c+1, c < c*2)
	println(c == -2228223, '\U0010ffff'*'\U0010ffff' > 2147483647)
	println(c > 1<<40, 'a'+1 == 'b')
}

// Output:
// true true true true
// false true
// true true
