package main

const (
	SmallestNonzeroFloat64 = 0x1p-1022 * 0x1p-52 // 4.9406564584124654417656879286822137236505980e-324
	DividedByTwo           = SmallestNonzeroFloat64 / 2
)

func main() {
	println(DividedByTwo)
}

// Output:
// 0
