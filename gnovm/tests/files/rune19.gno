package main

// An explicit conversion to an interface gives an untyped rune constant
// its default type, rune, so the value must fit in int32.

func main() {
	var x = interface{}('a' << 40)
	println(x)
}

// Error:
// main/rune19.gno:7:10-32: rune overflows target kind

// TypeCheckError:
// main/rune19.gno:7:22: 'a' << 40 (untyped rune constant 106652627894272) overflows rune
