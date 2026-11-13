package main

// Typed rune arithmetic keeps int32 semantics: it wraps at runtime.

func main() {
	var a rune = '\U0010ffff'
	println(a * a)
	var b int32 = 2147483647
	b++
	println(b, -b)
	var c rune = 'a'
	println(c<<30, c<<40)
	x := 'a' // defaults to rune
	println(x * 1000000000)
}

// Output:
// -2228223
// -2147483648 -2147483648
// 1073741824 0
// -1784247808
