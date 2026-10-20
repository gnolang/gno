package main

// Non-constant shifts of untyped rune constants: unary operators, two
// shifts of the same kind, explicit typed contexts and switch tags.

var s uint

const big = '\U0010ffff' * '\U0010ffff'

func main() {
	e := -('a' << s) // rune
	f := ('a' << s) + ('b' << s)
	_ = []rune{e, f}
	var i int64 = ('a' << s) + (1 << 40)
	var j int64 = (1 << 40) + ('a' << s)
	var k uint64 = 1 + ('a' << s)
	var l int64 = big << s
	var m int64 = ('a' << s) + big
	println(e, f, i, j, k, l, m)
	println(('a'<<s) == 97, ('a'<<s) == 'a', ('a'<<s) != big-big)

	switch +('a' << s) {
	case 'a':
		println("+ match")
	}
	switch -('a' << s) {
	case -'a':
		println("- match")
	}
	switch ^('a' << s) {
	case ^'a':
		println("^ match")
	}
}

// Output:
// -97 195 1099511627873 1099511627873 98 1241243320321 1241243320418
// true true true
// + match
// - match
// ^ match
