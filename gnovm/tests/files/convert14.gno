package main

func main() {
	b := []byte(1)
}

// Error:
// main/convert9.gno:4:7-16: cannot convert (const (1 <untyped> bigint)) to type []uint8

// TypeCheckError:
// main/convert9.gno:4:14: cannot convert 1 (untyped int constant) to type []byte; main/convert9.gno:4:2: declared and not used: b
