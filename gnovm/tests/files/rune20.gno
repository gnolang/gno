package main

// A non-constant shift of an untyped constant takes the type the constant
// would take alone; mixed with another untyped operand, both get the
// default type of their combined kind (integer < rune < float). Two
// non-constant operands of different kinds, ('a' << s) + (1 << s), take
// their type from context, which is not resolved here: see rune26-28.

var s uint

func main() {
	a := 1 + ('a' << s) // rune
	b := ('a' << s) + 1 // rune
	c := (1 << s) + 'a' // rune
	d := ('a' << s) * 2 // rune
	e := -('a' << s)    // rune
	f := ('a' << s) + ('b' << s)
	h := (1 << s) + (1 << 40) // int
	println([]rune{a, b, c, d, e, f}[5], a, b, c, d, e, f, h)

	// Comparisons.
	println(('a'<<s) < 100, 100 > ('a'<<s), ('a'<<s) == 97, ('a'<<s) == 'a')
	println(('a'<<s) < (1<<s), (1<<s) == ('a'<<s))

	// An explicit context type applies to the whole expression.
	var i int64 = ('a' << s) + (1 << 40)
	var j int64 = (1 << 40) + ('a' << s)
	var k uint64 = 1 + ('a' << s)
	var l int64 = ('\U0010ffff' * '\U0010ffff') << s
	println(i, j, k, l)
}

// Output:
// 195 98 98 98 194 -97 195 1099511627777
// true true true true
// false false
// 1099511627873 1099511627873 98 1241243320321
