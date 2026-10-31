package main

// Within the limits a constant may use all 512 bits and a shift count
// of 1074.

const c = 1 << 511

func main() {
	var x uint64 = c >> 448
	var y uint64 = (c - 1 + c) >> 450
	println(x, y, 0<<1074, c/2*2 == c, 'a'<<500>>500)
}

// Output:
// 9223372036854775808 4611686018427387903 0 true 97
