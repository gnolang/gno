package main

// A constant shift of a typed constant must be representable in the type;
// an untyped rune shifted past int32 stays exact until it takes a type.

const k uint = 3

func main() {
	var x int64 = 'a' << 40
	println(x)
	println(k << 63)
}

// Error:
// main/const_shift_overflow.gno:11:10-17: k << 63 (constant 27670116110564327424 of type uint) overflows uint

// TypeCheckError:
// main/const_shift_overflow.gno:11:10: k << 63 (constant 27670116110564327424 of type uint) overflows uint
