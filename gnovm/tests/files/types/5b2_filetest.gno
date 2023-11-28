package main

// one untyped const, one typed const
func main() {
	println(1 % "a")
}

// Error:
// main/files/types/5b2_filetest.gno:5: QUO or REM should have operand number, while have: <untyped> string
