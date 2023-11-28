package main

// both untyped const
// TODO: dec value representation
func main() {
	println(1.0 % 0)
}

// Error:
// main/files/types/5d4_filetest.gno:6: QUO or REM should have operand number, while have: <untyped> bigdec
