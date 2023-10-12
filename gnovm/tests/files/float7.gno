package main

const (
	SmallestNonzeroFloat32 = 0x1p-126 * 0x1p-23 // 1.401298464324817070923729583289916131280e-45
	DividedByTwo           = SmallestNonzeroFloat32 / 2
)

func main() {
	var i float32
	i = DividedByTwo
	println(i)
}

// Output:
// 0
