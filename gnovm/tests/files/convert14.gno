package main

// const -> slice
func main() {
	b := []byte(1)
}

// Error:
// main/convert14.gno:5:7-16: cannot convert (const (1 <untyped> bigint)) to type []uint8

// TypeCheckError:
// main/convert14.gno:5:14: cannot convert 1 (untyped int constant) to type []byte; main/convert14.gno:5:2: declared and not used: b
