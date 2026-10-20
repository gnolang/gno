package main

// string(x) of an integer constant outside the int32 range is "\uFFFD",
// however large the constant.

type S string

const r = string('\x01' << 64)

func main() {
	println(r == "\uFFFD")
	println(string(-'\x01'<<64) == "\uFFFD")
	println(S('\x01'<<64) == "\uFFFD")
	println(string(1<<64) == "\uFFFD", string(-1<<70) == "\uFFFD")
	println(string('\x01'<<40) == "\uFFFD", string('\x01'<<31) == "\uFFFD")
	println(string('a'+1), string(S('a'*2-'a')))
}

// Output:
// true
// true
// true
// true true
// true true
// b a
