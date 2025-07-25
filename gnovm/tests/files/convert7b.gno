package main

func main() {
	println(int64(1.5))
}

// Error:
// main/convert7b.gno:4:10-20: cannot convert (const (1.5 <untyped> bigdec)) to integer type

// TypeCheckError:
// main/convert7b.gno:4:16: cannot convert 1.5 (untyped float constant) to type int64
