package main

// A non-constant shift whose untyped rune operand is typed from its
// context shifts the exact constant (#6389); without a context it
// defaults to rune and overflows (rune_const_shift_context1.gno). The
// contexts #6393 lists are unchanged.

const c = '\U0010ffff' * '\U0010ffff'

var s uint = 1

func main() {
	var x int64 = c << s
	println(x, int64(c<<s), int64(c>>s))
}

// Output:
// 2482486640642 2482486640642 620621660160
