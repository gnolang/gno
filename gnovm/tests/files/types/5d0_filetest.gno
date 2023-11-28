package main

// both untyped const
// TODO: dec value representation, and this should happen in process stage!!!
func main() {
	println(1.0 % 1)
}

// Error:
// main/files/types/5d0_filetest.gno:6: QUO or REM should have operand number, while have: <untyped> bigdec
