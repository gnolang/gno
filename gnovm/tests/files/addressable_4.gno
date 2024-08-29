package main

type S struct {
	s string
}

func main() {
	// Strings are special; they are slicable even if not addressable.
	println("hello"[:])
	println("hello"[0])
	println("hello"[2:4])
	println(S{s: "hello"}.s[:])
	println(S{s: "hello"}.s[0])
	println(S{s: "hello"}.s[2:4])
}

// Output:
// hello
// 104
// ll
// hello
// 104
// ll
