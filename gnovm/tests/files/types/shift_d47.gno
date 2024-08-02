package main

var specialBytes [16]byte

func main() {
	for i, b := range []byte(`\.+*?()|[]{}^$`) {
		specialBytes[b%16] |= 1 << (b / 16)
		println(i, (1 << (b / 16)), specialBytes[b%16])
	}
}

// Output:
// 0 32 32
// 1 4 4
// 2 4 4
// 3 4 4
// 4 8 8
// 5 4 4
// 6 4 4
// 7 128 160
// 8 32 36
// 9 32 32
// 10 128 164
// 11 128 160
// 12 32 36
// 13 4 4
