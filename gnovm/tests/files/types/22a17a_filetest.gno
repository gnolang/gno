package main

type Int int8

func Add(a, b Int) Int {
	return a + b
}

// No binaryExpr or unaryExpr, no op related, still need checkType
func main() {
	println(Add(0, int8(1)))
}

// Error:
// main/files/types/22a17a_filetest.gno:11: cannot use int8 as main.Int without explicit conversion
