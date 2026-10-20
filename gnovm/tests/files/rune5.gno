package main

// Untyped rune constant declarations keep their exact value.

const (
	c   = '\U0010ffff' * '\U0010ffff' // beyond int32
	d   = c / '\U0010ffff'            // back in range: 1114111
	e   = c - c + 'a'                 // 97
	big = c * c * c                   // beyond int64, still a valid constant
)

func main() {
	var x int64 = c
	println(x)

	var r rune = d
	s := []rune{d, e}
	println(r, s[0], s[1], string(e))

	var y int64 = big / c / c
	println(y)

	// string of an out-of-range rune constant is "\uFFFD".
	println(string(c) == "\uFFFD")
}

// Output:
// 1241243320321
// 1114111 1114111 97 a
// 1241243320321
// true
