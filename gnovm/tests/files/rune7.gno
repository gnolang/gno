package main

// An operation between an untyped rune and an untyped integer constant
// yields an untyped rune (gnolang/gno#6339).

var (
	r1 = 'a' + 1
	r2 = 1 + 'a'
	r3 = 'a' * 2
	r4 = 'a' << 1
	rs = []rune{r1, r2, r3, r4}
)

const big = 'a' + 1<<40 // untyped rune beyond int32

func main() {
	println(rs[0], rs[1], rs[2], rs[3])
	var x int64 = big
	var y int64 = 1<<40 + 'a'
	println(x, y)
	r := big - 1<<40 // back in range, defaults to rune
	println([]rune{r}[0])
	// with an untyped float the result is an untyped float.
	var f float64 = 'a' + 0.5
	println(f == 97.5, int64(f*2))
}

// Output:
// 98 98 194 194
// 1099511627873 1099511627873
// 97
// true 195
