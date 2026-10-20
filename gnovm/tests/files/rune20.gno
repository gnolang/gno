package main

// Non-constant shifts of untyped rune constants: unary operators, two
// shifts of the same kind, explicit typed contexts and switch tags. An
// untyped rune constant beyond int32 next to a non-constant operand is
// rejected (rune7, rune21, rune23, rune27-rune33).

var s uint

func main() {
	e := -('a' << s) // rune
	f := ('a' << s) + ('b' << s)
	_ = []rune{e, f}
	var i int64 = ('a' << s) + (1 << 40)
	var j int64 = (1 << 40) + ('a' << s)
	var k uint64 = 1 + ('a' << s)
	println(e, f, i, j, k)
	println(('a'<<s) == 97, ('a'<<s) == 'a', ('a'<<s) != 0)

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
	switch ('a' << s) + 'a' {
	case 'a' + 'a':
		println("compound match")
	}
}

// Output:
// -97 195 1099511627873 1099511627873 98
// true true true
// + match
// - match
// ^ match
// compound match
