package main

// In an interface context an untyped expression takes its default type as
// a whole: 1 + ('a' << s) is a rune, not an int plus a rune.

var s uint

func f(x interface{}) interface{} { return x }

func main() {
	var x interface{} = 1 + ('a' << s)
	_, ok1 := x.(rune)
	_, ok2 := f(1 + ('a' << s)).(rune)
	_, ok3 := interface{}(1 + ('a' << s)).(rune)
	m := map[interface{}]int{}
	m[1+('a'<<s)] = 1
	_, ok4 := m[rune('b')]
	println(ok1, ok2, ok3, ok4)

	var y interface{} = 1 + ((-'a') << s)
	println(y.(rune), 1+((-'a')<<s))
}

// Output:
// true true true true
// -96 -96
