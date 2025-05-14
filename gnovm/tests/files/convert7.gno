package main

func main() {
	println(int32(1.5))
}

// Error:
// main/files/convert7.gno:4:10: cannot convert (const (1.5 <untyped> bigdec)) to integer type

// TypeCheckError:
// main/files/convert7.gno:4:16: cannot convert 1.5 (untyped float constant) to type int32
