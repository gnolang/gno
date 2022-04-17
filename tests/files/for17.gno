package main

func main() {
	for i, chr := range "thisisastring" {
		println(i, string(chr))
	}
}

// Output:
// 0 t
// 1 h
// 2 i
// 3 s
// 4 i
// 5 s
// 6 a
// 7 s
// 8 t
// 9 r
// 10 i
// 11 n
// 12 g
