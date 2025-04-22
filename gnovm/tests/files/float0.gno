package main

func main() {
	x := int(1.2)
	println(x)
}

// Error:
// main/files/float0.gno:4:7: cannot convert (const (1.2 <untyped> bigdec)) to integer type

// TypeCheckError:
// main/files/float0.gno:4:11: cannot convert 1.2 (untyped float constant) to type int
