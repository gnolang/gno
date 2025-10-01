package main

func main() {
	x := int(1.2)
	println(x)
}

// Error:
// main/float0.gno:4:7-15: cannot convert (const (1.2 <untyped> bigdec)) to integer type

// TypeCheckError:
// main/float0.gno:4:11: cannot convert 1.2 (untyped float constant) to type int
