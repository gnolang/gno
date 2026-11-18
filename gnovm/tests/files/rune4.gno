package main

// Untyped rune constant arithmetic within int32 is exact and keeps the rune
// kind: every arithmetic, bitwise and shift operator.

const max = '\U0010ffff'

func main() {
	r := []rune{
		max * 'Ā', max + max, 'a' - max, max / 'a', max % 'a',
		max & 'a', max | 'a', max ^ 'a', max &^ 'a',
		'a' << 20, max >> 4,
		('\x00' - '\x01') << 31, // math.MinInt32
	}
	for _, v := range r {
		println(v)
	}
	println(max*'Ā' > 0, 'a' < 'b', max == 1114111, '\x00'-'a' < '\x00'-'\x01')
}

// Output:
// 285212416
// 2228222
// -1114014
// 11485
// 66
// 97
// 1114111
// 1114014
// 1114014
// 101711872
// 69631
// -2147483648
// true true true true
