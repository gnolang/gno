package main

// Untyped rune constants convert to float types, and string(x) of an
// integer constant beyond int64 is "\uFFFD".

type S string

func main() {
	var f float64 = 'a'
	var g float32 = ('\x00' - 'a') * 2
	println(int64(f), int64(g))

	println(string(1<<64) == "\uFFFD", string(-1<<70) == "\uFFFD", S(1<<64) == "\uFFFD")
	println(string(1<<40) == "\uFFFD", string(1<<31) == "\uFFFD")
	println(string('a'+1), string(S('a'*2-'a')))
}

// Output:
// 97 -194
// true true true
// true true
// b a
