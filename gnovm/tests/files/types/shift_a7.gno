package main

func main() {
	r := int(1)
	println(r << "a")
}

// Error:
// main/shift_a7.gno:5:10-18: cannot convert (const ("a" <untyped> string)) to type uint

// TypeCheckError:
// main/shift_a7.gno:5:15: cannot convert "a" (untyped string constant) to type uint
