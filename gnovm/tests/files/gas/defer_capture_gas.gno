package main

func main() {
	a, b, c, d := 1, 2, 3, 4
	e, f, g, h := 5, 6, 7, 8
	for i := 0; i < 10; i++ {
		defer func(x, y int) {
			_ = a + b + c + d + e + f + g + h + x + y
		}(i, i)
	}
}

// Gas:
// 69271
