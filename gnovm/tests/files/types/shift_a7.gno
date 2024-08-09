package main

func main() {
	r := int(1)
	println(r << "a")
}

// Error:
// main/files/types/shift_a7.gno:5:10: invalid operation: invalid shift count: (const ("a" <untyped> string))
