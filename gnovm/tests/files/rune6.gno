package main

// Shifts of an untyped rune constant are exact.

const shl = 'a' << 40

func main() {
	var a int64 = 'a' << 40
	var b int64 = shl
	var c uint64 = '\U0010ffff' << 43
	var r rune = shl >> 38 // back in range
	println(a, b, c, []rune{r}[0])
	var d int64 = ('\U0010ffff' * '\U0010ffff') >> 1
	println(d)
}

// Output:
// 106652627894272 106652627894272 9799823993065177088 388
// 620621660160
